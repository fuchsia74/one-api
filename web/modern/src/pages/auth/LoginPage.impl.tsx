import { useEffect, useRef, useState } from "react";
import {
  useNavigate,
  Link,
  useSearchParams,
  useLocation,
} from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import * as z from "zod";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from "@/components/ui/form";
import { Separator } from "@/components/ui/separator";
import { useAuthStore } from "@/lib/stores/auth";
import { api } from "@/lib/api";

const loginSchema = z.object({
  username: z.string().min(1, "Username is required"),
  password: z.string().min(1, "Password is required"),
  totp_code: z
    .string()
    .optional()
    .refine((val) => !val || val.length === 6, {
      message: "TOTP code must be 6 digits",
    }),
});

type LoginForm = z.infer<typeof loginSchema>;

interface SystemStatus {
  github_oauth?: boolean;
  github_client_id?: string;
  wechat_login?: boolean;
  lark_client_id?: string;
  system_name?: string;
  logo?: string;
}

export function LoginPage() {
  const [isLoading, setIsLoading] = useState(false);
  const [totpRequired, setTotpRequired] = useState(false);
  const [systemStatus, setSystemStatus] = useState<SystemStatus>({});
  const [successMessage, setSuccessMessage] = useState<string>("");
  const [totpValue, setTotpValue] = useState("");
  const totpRef = useRef<HTMLInputElement | null>(null);
  const navigate = useNavigate();
  const [searchParams] = useSearchParams();
  const location = useLocation();
  const { login } = useAuthStore();

  const form = useForm<LoginForm>({
    resolver: zodResolver(loginSchema),
    defaultValues: { username: "", password: "", totp_code: "" },
  });

  useEffect(() => {
    // Check for expired session
    if (searchParams.get("expired")) {
      console.warn("Session expired, please login again");
    }

    // Handle success messages from navigation state
    if (location.state?.message) {
      setSuccessMessage(location.state.message);
      // Clear the state to prevent showing the message on refresh
      window.history.replaceState({}, document.title);
    }

    // Load system status
    const status = localStorage.getItem("status");
    if (status) {
      try {
        setSystemStatus(JSON.parse(status));
      } catch (error) {
        console.error("Error parsing system status:", error);
      }
    }
  }, [searchParams, location.state]);

  const onGitHubOAuth = () => {
    if (systemStatus.github_client_id) {
      const redirectUri = `${window.location.origin}/oauth/github`;
      window.location.href = `https://github.com/login/oauth/authorize?client_id=${systemStatus.github_client_id}&redirect_uri=${redirectUri}&scope=user:email`;
    }
  };

  const onLarkOAuth = () => {
    if (systemStatus.lark_client_id) {
      const redirectUri = `${window.location.origin}/oauth/lark`;
      window.location.href = `https://open.larksuite.com/open-apis/authen/v1/index?app_id=${systemStatus.lark_client_id}&redirect_uri=${redirectUri}`;
    }
  };

  const onSubmit = async (data: LoginForm) => {
    setIsLoading(true);
    try {
      const payload: Record<string, string> = {
        username: data.username,
        password: data.password,
      };
      if (totpRequired && totpValue) payload.totp_code = totpValue;
      // Unified API call - complete URL with /api prefix
      const response = await api.post("/api/user/login", payload);
      const { success, message, data: respData } = response.data;
      const m = typeof message === "string" ? message.trim().toLowerCase() : "";
      const dataTotp = !!(
        respData &&
        (respData.totp_required === true ||
          respData.totp_required === "true" ||
          respData.totp_required === 1)
      );
      const needsTotp =
        !success && (dataTotp || m === "totp_required" || m.includes("totp"));

      if (needsTotp) {
        setTotpRequired(true);
        setTotpValue("");
        form.setValue("totp_code", "");
        form.setError("root", { message: "Please enter your TOTP code" });
        return;
      }

      if (success) {
        login(respData, "");

        // Get redirect_to parameter from URL
        const redirectTo = searchParams.get("redirect_to");

        // Handle default root password warning
        if (data.username === "root" && data.password === "123456") {
          navigate("/users/edit");
          console.warn("Please change the default root password");
        } else if (redirectTo) {
          // Decode and navigate to the original page
          try {
            const decodedPath = decodeURIComponent(redirectTo);
            // Ensure the redirect path is safe (starts with /)
            if (decodedPath.startsWith("/")) {
              navigate(decodedPath);
            } else {
              navigate("/dashboard");
            }
          } catch (error) {
            console.error("Invalid redirect_to parameter:", error);
            navigate("/dashboard");
          }
        } else {
          navigate("/dashboard");
        }
      } else {
        form.setError("root", {
          message:
            m === "totp_required"
              ? "Please enter your TOTP code"
              : message || "Login failed",
        });
      }
    } catch (error) {
      form.setError("root", {
        message: error instanceof Error ? error.message : "Login failed",
      });
    } finally {
      setIsLoading(false);
    }
  };

  useEffect(() => {
    if (totpRequired && totpRef.current) totpRef.current.focus();
  }, [totpRequired]);

  const hasOAuthOptions =
    systemStatus.github_oauth ||
    systemStatus.wechat_login ||
    systemStatus.lark_client_id;

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <Card className="w-full max-w-md">
        <CardHeader className="text-center">
          {systemStatus.logo && (
            <div className="flex justify-center mb-4">
              <img src={systemStatus.logo} alt="Logo" className="h-12 w-auto" />
            </div>
          )}
          <CardTitle className="text-2xl">
            Sign In
            {systemStatus.system_name ? ` to ${systemStatus.system_name}` : ""}
          </CardTitle>
          <CardDescription>
            Enter your credentials to access your account
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Form {...form}>
            <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
              <FormField
                control={form.control}
                name="username"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Username</FormLabel>
                    <FormControl>
                      <Input {...field} disabled={totpRequired} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={form.control}
                name="password"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Password</FormLabel>
                    <FormControl>
                      <Input
                        type="password"
                        {...field}
                        disabled={totpRequired}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              {totpRequired && (
                <FormField
                  control={form.control}
                  name="totp_code"
                  render={({ field }) => (
                    <FormItem>
                      <FormLabel>TOTP Code</FormLabel>
                      <FormControl>
                        <Input
                          maxLength={6}
                          placeholder="Enter 6-digit TOTP code"
                          {...field}
                          ref={totpRef}
                          inputMode="numeric"
                          pattern="[0-9]*"
                          onChange={(e) => {
                            field.onChange(e);
                            setTotpValue(e.target.value);
                          }}
                        />
                      </FormControl>
                      <FormMessage />
                    </FormItem>
                  )}
                />
              )}
              {successMessage && (
                <div className="text-sm text-green-600 bg-green-50 p-3 rounded-md border border-green-200">
                  {successMessage}
                </div>
              )}
              {form.formState.errors.root && (
                <div className="text-sm text-destructive">
                  {totpRequired
                    ? "Please enter your TOTP code"
                    : form.formState.errors.root.message}
                </div>
              )}
              <Button
                type="submit"
                className="w-full"
                disabled={isLoading || (totpRequired && totpValue.length !== 6)}
              >
                {isLoading
                  ? "Signing in..."
                  : totpRequired
                    ? "Verify TOTP"
                    : "Sign In"}
              </Button>

              {totpRequired && (
                <Button
                  type="button"
                  variant="outline"
                  className="w-full"
                  onClick={() => {
                    setTotpRequired(false);
                    setTotpValue("");
                    form.setValue("totp_code", "");
                    form.clearErrors("root");
                  }}
                >
                  Back to Login
                </Button>
              )}

              <div className="text-center text-sm space-y-2">
                <Link to="/reset" className="text-primary hover:underline">
                  Forgot your password?
                </Link>
                <div>
                  Don't have an account?{" "}
                  <Link to="/register" className="text-primary hover:underline">
                    Sign up
                  </Link>
                </div>
              </div>
            </form>
          </Form>

          {hasOAuthOptions && (
            <>
              <Separator className="my-4" />
              <div className="text-center">
                <p className="text-sm text-muted-foreground mb-4">
                  Or continue with
                </p>
                <div className="flex justify-center gap-2">
                  {systemStatus.github_oauth && (
                    <Button variant="outline" size="sm" onClick={onGitHubOAuth}>
                      <svg
                        className="w-4 h-4 mr-2"
                        viewBox="0 0 24 24"
                        fill="currentColor"
                      >
                        <path d="M12 0c-6.626 0-12 5.373-12 12 0 5.302 3.438 9.8 8.207 11.387.599.111.793-.261.793-.577v-2.234c-3.338.726-4.033-1.416-4.033-1.416-.546-1.387-1.333-1.756-1.333-1.756-1.089-.745.083-.729.083-.729 1.205.084 1.839 1.237 1.839 1.237 1.07 1.834 2.807 1.304 3.492.997.107-.775.418-1.305.762-1.604-2.665-.305-5.467-1.334-5.467-5.931 0-1.311.469-2.381 1.236-3.221-.124-.303-.535-1.524.117-3.176 0 0 1.008-.322 3.301 1.23.957-.266 1.983-.399 3.003-.404 1.02.005 2.047.138 3.006.404 2.291-1.552 3.297-1.23 3.297-1.23.653 1.653.242 2.874.118 3.176.77.84 1.235 1.911 1.235 3.221 0 4.609-2.807 5.624-5.479 5.921.43.372.823 1.102.823 2.222v3.293c0 .319.192.694.801.576 4.765-1.589 8.199-6.086 8.199-11.386 0-6.627-5.373-12-12-12z" />
                      </svg>
                      GitHub
                    </Button>
                  )}
                  {systemStatus.wechat_login && (
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={() =>
                        console.log("WeChat OAuth not implemented")
                      }
                    >
                      <svg
                        className="w-4 h-4 mr-2"
                        viewBox="0 0 24 24"
                        fill="currentColor"
                      >
                        <path d="M8.691 2.188C3.891 2.188 0 5.476 0 9.53c0 2.212 1.146 4.203 2.943 5.652-.171.171-.684 1.026-.684 1.026s.342.171.684 0c.342-.171 1.368-.684 1.539-.855 1.368.342 2.736.513 4.209.513.342 0 .684 0 1.026-.171-.171-.342-.171-.684-.171-1.026 0-3.55 3.038-6.417 6.759-6.417.513 0 .855 0 1.368.171C16.187 4.741 12.809 2.188 8.691 2.188zM6.297 7.701c-.513 0-.855-.513-.855-1.026s.342-1.026.855-1.026c.513 0 .855.513.855 1.026s-.342 1.026-.855 1.026zm4.55 0c-.513 0-.855-.513-.855-1.026s.342-1.026.855-1.026c.513 0 .855.513.855 1.026s-.342 1.026-.855 1.026z" />
                        <path d="M15.733 9.36c-3.721 0-6.588 2.526-6.588 5.652 0 3.125 2.867 5.652 6.588 5.652 1.197 0 2.394-.342 3.42-.855.342.171 1.026.513 1.368.684.171.171.513 0 .513 0s-.342-.684-.513-1.026c1.539-1.197 2.526-2.867 2.526-4.721 0-3.125-2.867-5.652-6.588-5.652zM13.852 13.422c-.342 0-.684-.342-.684-.684s.342-.684.684-.684c.342 0 .684.342.684.684s-.342.684-.684.684zm3.42 0c-.342 0-.684-.342-.684-.684s.342-.684.684-.684c.342 0 .684.342.684.684s-.342.684-.684.684z" />
                      </svg>
                      WeChat
                    </Button>
                  )}
                  {systemStatus.lark_client_id && (
                    <Button variant="outline" size="sm" onClick={onLarkOAuth}>
                      <svg
                        className="w-4 h-4 mr-2"
                        viewBox="0 0 24 24"
                        fill="currentColor"
                      >
                        <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z" />
                      </svg>
                      Lark
                    </Button>
                  )}
                </div>
              </div>
            </>
          )}
        </CardContent>
      </Card>
    </div>
  );
}

export default LoginPage;
