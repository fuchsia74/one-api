package main

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Laisky/errors/v2"
	glog "github.com/Laisky/go-utils/v5/log"
	"github.com/Laisky/zap"
	"golang.org/x/sync/errgroup"
)

// run orchestrates the regression sweep across the configured models and variants.
func run(ctx context.Context, logger glog.Logger) error {
	cfg, err := loadConfig()
	if err != nil {
		return errors.Wrap(err, "load config")
	}

	variantLabels := make([]string, 0, len(cfg.Variants))
	for _, v := range cfg.Variants {
		variantLabels = append(variantLabels, v.Header)
	}
	logger.Info("starting API regression sweep",
		zap.String("base_url", cfg.APIBase),
		zap.Int("model_count", len(cfg.Models)),
		zap.Int("variant_count", len(cfg.Variants)),
		zap.Strings("variants", variantLabels),
	)

	httpClient := &http.Client{Timeout: 60 * time.Second}
	resultsCh := make(chan testResult, len(cfg.Models)*len(cfg.Variants))

	var (
		results   []testResult
		collectWg sync.WaitGroup
	)
	collectWg.Add(1)
	go func() {
		defer collectWg.Done()
		for res := range resultsCh {
			results = append(results, res)
			switch {
			case res.Success:
				logger.Info("request succeeded",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Duration("duration", res.Duration),
					zap.Int("status", res.StatusCode),
				)
			case res.Skipped:
				logger.Info("request skipped",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Int("status", res.StatusCode),
					zap.String("reason", res.ErrorReason),
				)
			default:
				logger.Warn("request failed",
					zap.String("model", res.Model),
					zap.String("variant", res.Label),
					zap.String("type", string(res.Type)),
					zap.Bool("stream", res.Stream),
					zap.Duration("duration", res.Duration),
					zap.Int("status", res.StatusCode),
					zap.String("error", res.ErrorReason),
					zap.String("request_body", res.RequestBody),
					zap.String("response_body", res.ResponseBody),
				)
			}
		}
	}()

	grp, grpCtx := errgroup.WithContext(ctx)
	for _, modelName := range cfg.Models {
		model := modelName
		grp.Go(func() error {
			executeModelSweep(grpCtx, httpClient, cfg, model, resultsCh)
			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		close(resultsCh)
		collectWg.Wait()
		return errors.Wrap(err, "await model sweeps")
	}

	close(resultsCh)
	collectWg.Wait()

	report := buildReport(cfg.Models, cfg.Variants, results)
	renderReport(report)

	if report.failedCount > 0 {
		return errors.Errorf("%d of %d requests failed", report.failedCount, report.totalRequests)
	}

	return nil
}

// executeModelSweep runs all variants for a particular model and publishes the results.
func executeModelSweep(ctx context.Context, client *http.Client, cfg config, model string, results chan<- testResult) {
	specs := buildRequestSpecs(model, cfg.Variants)

	innerGrp, innerCtx := errgroup.WithContext(ctx)
	for _, spec := range specs {
		s := spec
		if skip, reason := shouldSkipVariant(model, s); skip {
			outcome := testResult{
				Model:       model,
				Variant:     s.Variant,
				Label:       s.Label,
				Type:        s.Type,
				Stream:      s.Stream,
				Skipped:     true,
				ErrorReason: reason,
			}
			select {
			case results <- outcome:
			case <-innerCtx.Done():
			}
			continue
		}
		innerGrp.Go(func() error {
			res := performRequest(innerCtx, client, cfg.APIBase, cfg.Token, s, model)
			select {
			case results <- res:
			case <-innerCtx.Done():
			}
			return nil
		})
	}

	_ = innerGrp.Wait()
}

// buildRequestSpecs constructs the concrete request payloads for each variant.
func buildRequestSpecs(model string, variants []requestVariant) []requestSpec {
	specs := make([]requestSpec, 0, len(variants))
	for _, variant := range variants {
		var body any
		switch variant.Type {
		case requestTypeChatCompletion:
			body = chatCompletionPayload(model, variant.Stream, variant.Expectation)
		case requestTypeResponseAPI:
			body = responseAPIPayload(model, variant.Stream, variant.Expectation)
		case requestTypeClaudeMessages:
			body = claudeMessagesPayload(model, variant.Stream, variant.Expectation)
		}
		specs = append(specs, requestSpec{
			Variant:     variant.Key,
			Label:       variant.Header,
			Type:        variant.Type,
			Path:        variant.Path,
			Body:        body,
			Stream:      variant.Stream,
			Expectation: variant.Expectation,
		})
	}

	return specs
}

// shouldSkipVariant reports whether the provided request specification should be skipped for the model.
// The second return value describes the reason when the combination is unsupported.
func shouldSkipVariant(model string, spec requestSpec) (bool, string) {
	if spec.Expectation == expectationStructuredOutput {
		if reasons, ok := structuredVariantSkips[spec.Variant]; ok {
			if reason, exists := reasons[strings.ToLower(model)]; exists {
				return true, reason
			}
		}
	}

	if spec.Expectation != expectationVision {
		return false, ""
	}

	lower := strings.ToLower(model)
	if _, unsupported := visionUnsupportedModels[lower]; unsupported {
		return true, "vision input unsupported by model " + model
	}

	return false, ""
}
