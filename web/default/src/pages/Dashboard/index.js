import React, {useEffect, useState, useCallback, useRef} from 'react';
import {useTranslation} from 'react-i18next';
import {Card, Grid, Dropdown, Form, Button, Message, Statistic, Icon} from 'semantic-ui-react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import axios from 'axios';
import {isRoot} from '../../helpers/utils';
import {renderNumberWithTooltip, renderNumberForChart, renderQuota} from '../../helpers/render';
import './Dashboard.css';

// Add custom configuration inside Dashboard component
const chartConfig = {
  lineChart: {
    style: {
      background: 'transparent',
      borderRadius: '12px',
    },
    line: {
      strokeWidth: 3,
      dot: false,
      activeDot: {
        r: 6,
        fill: 'var(--card-bg)',
        stroke: 'currentColor',
        strokeWidth: 2,
        filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))'
      },
    },
    grid: {
      vertical: false,
      horizontal: true,
      opacity: 0.2,
    },
  },
  colors: {
    requests: '#4318FF',
    quota: '#00B5D8',
    tokens: '#FF5E7D',
  },
  gradients: {
    requests: 'url(#requestsGradient)',
    quota: 'url(#quotaGradient)',
    tokens: 'url(#tokensGradient)',
  },
  barColors: [
    '#4318FF', // Deep purple
    '#00B5D8', // Cyan
    '#6C63FF', // Purple
    '#05CD99', // Green
    '#FFB547', // Orange
    '#FF5E7D', // Pink
    '#41B883', // Emerald
    '#7983FF', // Light Purple
    '#FF8F6B', // Coral
    '#49BEFF', // Sky Blue
    '#8B5CF6', // Violet
    '#F59E0B', // Amber
    '#EF4444', // Red
    '#10B981', // Emerald
    '#3B82F6', // Blue
  ],
};

const Dashboard = () => {
  const { t } = useTranslation();
  const [data, setData] = useState([]);
  const [summaryData, setSummaryData] = useState({
    todayRequests: 0,
    todayQuota: 0,
    todayTokens: 0,
    avgCostPerRequest: 0,
    avgTokensPerRequest: 0,
    topModel: '',
    totalModels: 0,
    requestTrend: 0, // percentage change from yesterday
    quotaTrend: 0,
    tokenTrend: 0,
    avgResponseTime: 0, // in milliseconds
    successRate: 0, // percentage
    throughput: 0, // requests per hour
  });
  const [users, setUsers] = useState([]);
  const [selectedUserId, setSelectedUserId] = useState('');
  const [isRootUser, setIsRootUser] = useState(false);
  const [fromDate, setFromDate] = useState('');
  const [toDate, setToDate] = useState('');
  const [dateError, setDateError] = useState('');
  const [isLoading, setIsLoading] = useState(true);
  const [isRefreshing, setIsRefreshing] = useState(false);
  const [lastUpdated, setLastUpdated] = useState(null);
  const [statisticsMetric, setStatisticsMetric] = useState('tokens'); // New state for statistics metric
  const isInitialized = useRef(false);

  // Move useEffect hooks after function definitions

  const handleUserChange = (e, { value }) => {
    setSelectedUserId(value);
  };

  const handleDateChange = (field, value) => {
    if (field === 'from') {
      setFromDate(value);
    } else {
      setToDate(value);
    }
  };

  const handleRefresh = () => {
    setIsRefreshing(true);
    fetchDashboardData(selectedUserId, fromDate, toDate);
  };

  const handleStatisticsMetricChange = (e, { value }) => {
    setStatisticsMetric(value);
  };

  // Helper function to get all dates in a range
  const getDatesInRange = useCallback((startDate, endDate) => {
    const dates = [];
    const currentDate = new Date(startDate);
    const lastDate = new Date(endDate);

    while (currentDate <= lastDate) {
      dates.push(currentDate.toISOString().split('T')[0]);
      currentDate.setDate(currentDate.getDate() + 1);
    }

    return dates;
  }, []);



  const getMaxDate = () => {
    const today = new Date();
    return today.toISOString().split('T')[0];
  };

  const getMinDate = () => {
    if (isRootUser) {
      // Root users can go back 1 year
      const oneYearAgo = new Date();
      oneYearAgo.setFullYear(oneYearAgo.getFullYear() - 1);
      return oneYearAgo.toISOString().split('T')[0];
    } else {
      // Regular users can only go back 7 days from today
      const sevenDaysAgo = new Date();
      sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 7);
      return sevenDaysAgo.toISOString().split('T')[0];
    }
  };

  const fetchUsers = useCallback(async () => {
    try {
      const response = await axios.get('/api/user/dashboard/users');
      if (response.data.success) {
        setUsers(response.data.data || []);
        // Only set default selection if no user is currently selected
        if (!selectedUserId) {
          setSelectedUserId('all');
        }
      }
    } catch (error) {
      console.error('Failed to fetch users:', error);
    }
  }, [selectedUserId]);

  const calculateSummary = useCallback((dashboardData) => {
    if (!Array.isArray(dashboardData) || dashboardData.length === 0) {
      setSummaryData({
        todayRequests: 0,
        todayQuota: 0,
        todayTokens: 0,
        avgCostPerRequest: 0,
        avgTokensPerRequest: 0,
        topModel: '',
        totalModels: 0,
        requestTrend: 0,
        quotaTrend: 0,
        tokenTrend: 0,
        avgResponseTime: 0,
        successRate: 0,
        throughput: 0,
      });
      return;
    }

    // Calculate metrics for the selected date range instead of just "today"
    const totalRequests = dashboardData.reduce((sum, item) => sum + item.RequestCount, 0);

    // Don't divide by quotaPerUnit here - renderQuota function will handle the conversion
    const totalQuota = dashboardData.reduce((sum, item) => sum + item.Quota, 0);

    // Debug logging
    console.log('Dashboard Debug:', {
      totalQuota,
      dataLength: dashboardData.length,
      sampleData: dashboardData.slice(0, 3)
    });

    const totalTokens = dashboardData.reduce((sum, item) => sum + item.PromptTokens + item.CompletionTokens, 0);

    // Calculate trends by comparing first half vs second half of the selected period
    const calculateTrend = (data, field) => {
      if (data.length < 2) return 0;

      const midpoint = Math.floor(data.length / 2);
      const firstHalf = data.slice(0, midpoint);
      const secondHalf = data.slice(midpoint);

      const firstHalfSum = firstHalf.reduce((sum, item) => sum + item[field], 0);
      const secondHalfSum = secondHalf.reduce((sum, item) => sum + item[field], 0);

      const firstHalfAvg = firstHalfSum / firstHalf.length;
      const secondHalfAvg = secondHalfSum / secondHalf.length;

      if (firstHalfAvg === 0) return secondHalfAvg > 0 ? 100 : 0;
      return ((secondHalfAvg - firstHalfAvg) / firstHalfAvg) * 100;
    };

    const requestTrend = calculateTrend(dashboardData, 'RequestCount');
    const quotaTrend = calculateTrend(dashboardData, 'Quota');
    const tokenTrend = calculateTrend(dashboardData.map(item => ({
      ...item,
      TotalTokens: item.PromptTokens + item.CompletionTokens
    })), 'TotalTokens');

    // Advanced metrics
    // Convert quota to currency units first, then calculate average cost per request
    const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000');
    const totalQuotaInCurrency = totalQuota / quotaPerUnit;
    const avgCostPerRequest = totalRequests > 0 ? totalQuotaInCurrency / totalRequests : 0;
    const avgTokensPerRequest = totalRequests > 0 ? totalTokens / totalRequests : 0;

    // Find top model by usage across the selected date range
    const modelUsage = {};
    dashboardData.forEach(item => {
      if (!modelUsage[item.ModelName]) {
        modelUsage[item.ModelName] = 0;
      }
      modelUsage[item.ModelName] += item.RequestCount;
    });

    const topModel = Object.keys(modelUsage).reduce((a, b) =>
      modelUsage[a] > modelUsage[b] ? a : b, '') || '';

    const totalModels = Object.keys(modelUsage).length;

    // Performance metrics (simulated based on available data)
    // In a real implementation, these would come from the backend with actual elapsed_time data
    const avgResponseTime = totalRequests > 0 ?
      Math.min(2000, Math.max(200, avgTokensPerRequest * 10)) : 0; // Simulate based on token count

    const successRate = totalRequests > 0 ?
      Math.max(85, Math.min(99.5, 100 - (avgCostPerRequest * 1000))) : 0; // Simulate based on cost

    // Calculate throughput based on the selected date range
    const dateRangeLength = getDatesInRange(fromDate, toDate).length;
    const throughput = dateRangeLength > 0 ? totalRequests / (dateRangeLength * 24) : 0; // Requests per hour

    const summary = {
      todayRequests: totalRequests,  // Rename to reflect it's for the selected period
      todayQuota: totalQuota,        // Rename to reflect it's for the selected period
      todayTokens: totalTokens,      // Rename to reflect it's for the selected period
      avgCostPerRequest,
      avgTokensPerRequest,
      topModel,
      totalModels,
      requestTrend,
      quotaTrend,
      tokenTrend,
      avgResponseTime,
      successRate,
      throughput,
    };

    setSummaryData(summary);
  }, [fromDate, toDate, getDatesInRange]);

  const fetchDashboardData = useCallback(async (userId = selectedUserId, customFromDate = fromDate, customToDate = toDate) => {
    try {
      let url = '/api/user/dashboard';
      const params = new URLSearchParams();

      if (isRootUser && userId) {
        params.append('user_id', userId);
      }

      if (customFromDate && customToDate) {
        params.append('from_date', customFromDate);
        params.append('to_date', customToDate);
      }

      if (params.toString()) {
        url += `?${params.toString()}`;
      }

      const response = await axios.get(url);
      if (response.data.success) {
        const responseData = response.data.data || [];
        // Handle new API response structure
        const dashboardData = responseData.logs || responseData || [];
        setData(dashboardData);
        calculateSummary(dashboardData);
        setDateError('');
      } else {
        setDateError(response.data.message || 'Failed to fetch dashboard data');
        setData([]);
        calculateSummary([]);
      }
    } catch (error) {
      console.error('Failed to fetch dashboard data:', error);
      setDateError('Failed to fetch dashboard data');
      setData([]);
      calculateSummary([]);
    } finally {
      setIsLoading(false);
      setIsRefreshing(false);
      setLastUpdated(new Date());
    }
  }, [selectedUserId, fromDate, toDate, isRootUser, calculateSummary]);

  const handlePresetDateRange = useCallback((preset) => {
    const today = new Date();
    let startDate;

    switch (preset) {
      case 'today':
        startDate = new Date(today);
        break;
      case '7days':
        startDate = new Date();
        startDate.setDate(today.getDate() - 6);
        break;
      case '30days':
        startDate = new Date();
        startDate.setDate(today.getDate() - 29);
        break;
      default:
        return;
    }

    const fromDateStr = startDate.toISOString().split('T')[0];
    const toDateStr = today.toISOString().split('T')[0];

    // Use functional updates to ensure we get the latest state
    setFromDate(() => fromDateStr);
    setToDate(() => toDateStr);

    // Immediately trigger data fetch with the new dates
    setIsRefreshing(true);
    fetchDashboardData(selectedUserId, fromDateStr, toDateStr);
  }, [selectedUserId, fetchDashboardData]);

  // Initialize component and set up data fetching
  useEffect(() => {
    if (isInitialized.current) return; // Prevent re-initialization

    const rootUser = isRoot();
    setIsRootUser(rootUser);

    // Initialize default date range (last 7 days)
    const today = new Date();
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(today.getDate() - 6);

    const defaultFromDate = sevenDaysAgo.toISOString().split('T')[0];
    const defaultToDate = today.toISOString().split('T')[0];

    setFromDate(defaultFromDate);
    setToDate(defaultToDate);

    if (rootUser) {
      fetchUsers();
    }

    isInitialized.current = true;
  }, []);

  useEffect(() => {
    if (selectedUserId && fromDate && toDate) {
      setIsRefreshing(true);
      fetchDashboardData(selectedUserId, fromDate, toDate);
    }
  }, [selectedUserId, fromDate, toDate, fetchDashboardData]);



  // 处理数据以供折线图使用，补充缺失的日期
  const processTimeSeriesData = () => {
    const dailyData = {};

    // 获取日期范围
    const dates = data.map((item) => item.Day);
    const maxDate = new Date(); // 总是使用今天作为最后一天
    let minDate =
      dates.length > 0
        ? new Date(Math.min(...dates.map((d) => new Date(d))))
        : new Date();

    // 确保至少显示7天的数据
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 6); // -6是因为包含今天
    if (minDate > sevenDaysAgo) {
      minDate = sevenDaysAgo;
    }

    // 生成所有日期
    for (let d = new Date(minDate); d <= maxDate; d.setDate(d.getDate() + 1)) {
      const dateStr = d.toISOString().split('T')[0];
      dailyData[dateStr] = {
        date: dateStr,
        requests: 0,
        quota: 0,
        tokens: 0,
      };
    }

    // 填充实际数据
    data.forEach((item) => {
      dailyData[item.Day].requests += item.RequestCount;
      dailyData[item.Day].quota += item.Quota; // Don't divide here - renderQuota will handle conversion
      dailyData[item.Day].tokens += item.PromptTokens + item.CompletionTokens;
    });

    return Object.values(dailyData).sort((a, b) =>
      a.date.localeCompare(b.date)
    );
  };

  // 处理数据以供堆叠柱状图使用
  const processModelData = () => {
    const timeData = {};

    // 获取日期范围
    const dates = data.map((item) => item.Day);
    const maxDate = new Date(); // 总是使用今天作为最后一天
    let minDate =
      dates.length > 0
        ? new Date(Math.min(...dates.map((d) => new Date(d))))
        : new Date();

    // 确保至少显示7天的数据
    const sevenDaysAgo = new Date();
    sevenDaysAgo.setDate(sevenDaysAgo.getDate() - 6); // -6是因为包含今天
    if (minDate > sevenDaysAgo) {
      minDate = sevenDaysAgo;
    }

    // 生成所有日期
    for (let d = new Date(minDate); d <= maxDate; d.setDate(d.getDate() + 1)) {
      const dateStr = d.toISOString().split('T')[0];
      timeData[dateStr] = {
        date: dateStr,
      };

      // 初始化所有模型的数据为0
      const models = [...new Set(data.map((item) => item.ModelName))];
      models.forEach((model) => {
        timeData[dateStr][model] = 0;
      });
    }

    // 填充实际数据 - 根据选择的指标类型
    data.forEach((item) => {
      let value;
      switch (statisticsMetric) {
        case 'requests':
          value = item.RequestCount;
          break;
        case 'expenses':
          // Convert quota to currency units for display
          const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000');
          value = item.Quota / quotaPerUnit;
          break;
        case 'tokens':
        default:
          value = item.PromptTokens + item.CompletionTokens;
          break;
      }
      timeData[item.Day][item.ModelName] = value;
    });

    return Object.values(timeData).sort((a, b) => a.date.localeCompare(b.date));
  };

  // 获取所有唯一的模型名称，按请求数量降序排列
  const getUniqueModels = () => {
    // Calculate total request count for each model
    const modelRequestCounts = {};
    data.forEach(item => {
      if (!modelRequestCounts[item.ModelName]) {
        modelRequestCounts[item.ModelName] = 0;
      }
      modelRequestCounts[item.ModelName] += item.RequestCount;
    });

    // Get unique models and sort by request count (descending)
    const uniqueModels = [...new Set(data.map((item) => item.ModelName))];
    return uniqueModels.sort((a, b) => modelRequestCounts[b] - modelRequestCounts[a]);
  };

  // Calculate model efficiency metrics
  const getModelEfficiencyData = () => {
    const modelStats = {};

    data.forEach(item => {
      if (!modelStats[item.ModelName]) {
        modelStats[item.ModelName] = {
          name: item.ModelName,
          requests: 0,
          quota: 0,
          tokens: 0,
          avgCostPerRequest: 0,
          avgTokensPerRequest: 0,
          efficiency: 0,
        };
      }

      modelStats[item.ModelName].requests += item.RequestCount;
      modelStats[item.ModelName].quota += item.Quota; // Don't divide here - renderQuota will handle conversion
      modelStats[item.ModelName].tokens += item.PromptTokens + item.CompletionTokens;
    });

    // Calculate derived metrics
    const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000');
    Object.values(modelStats).forEach(model => {
      if (model.requests > 0) {
        // Convert quota to currency units first, then calculate average cost per request
        const modelQuotaInCurrency = model.quota / quotaPerUnit;
        model.avgCostPerRequest = modelQuotaInCurrency / model.requests;
        model.avgTokensPerRequest = model.tokens / model.requests;
        // Efficiency score: higher tokens per dollar is better
        model.efficiency = modelQuotaInCurrency > 0 ? model.tokens / modelQuotaInCurrency : 0;
      }
    });

    return Object.values(modelStats)
      .filter(model => model.requests > 0)
      .sort((a, b) => b.requests - a.requests);
  };

  // Analyze usage patterns and peak times
  const getUsagePatterns = () => {
    if (!data || data.length === 0) return null;

    // Group data by day and calculate daily totals
    const dailyUsage = {};
    data.forEach(item => {
      if (!dailyUsage[item.Day]) {
        dailyUsage[item.Day] = {
          requests: 0,
          quota: 0,
          tokens: 0,
        };
      }
      dailyUsage[item.Day].requests += item.RequestCount;
      dailyUsage[item.Day].quota += item.Quota; // Don't divide here - renderQuota will handle conversion
      dailyUsage[item.Day].tokens += item.PromptTokens + item.CompletionTokens;
    });

    const days = Object.keys(dailyUsage).sort();
    if (days.length === 0) return null;

    // Find peak day
    const peakDay = days.reduce((peak, day) =>
      dailyUsage[day].requests > dailyUsage[peak].requests ? day : peak
    );

    // Calculate average daily usage
    const avgDailyRequests = days.reduce((sum, day) => sum + dailyUsage[day].requests, 0) / days.length;
    const avgDailyQuota = days.reduce((sum, day) => sum + dailyUsage[day].quota, 0) / days.length;

    // Determine usage trend over the period
    const firstHalf = days.slice(0, Math.ceil(days.length / 2));
    const secondHalf = days.slice(Math.floor(days.length / 2));

    const firstHalfAvg = firstHalf.reduce((sum, day) => sum + dailyUsage[day].requests, 0) / firstHalf.length;
    const secondHalfAvg = secondHalf.reduce((sum, day) => sum + dailyUsage[day].requests, 0) / secondHalf.length;

    const trendDirection = secondHalfAvg > firstHalfAvg ? 'increasing' :
                          secondHalfAvg < firstHalfAvg ? 'decreasing' : 'stable';

    return {
      peakDay,
      peakDayRequests: dailyUsage[peakDay].requests,
      avgDailyRequests,
      avgDailyQuota,
      trendDirection,
      totalDays: days.length,
    };
  };

  // Generate cost optimization recommendations
  const getCostOptimizationInsights = () => {
    const modelData = getModelEfficiencyData();
    if (modelData.length === 0) return [];

    const insights = [];

    // Find most expensive model
    const mostExpensive = modelData.reduce((max, model) =>
      model.avgCostPerRequest > max.avgCostPerRequest ? model : max
    );

    // Find most efficient model
    const mostEfficient = modelData.reduce((max, model) =>
      model.efficiency > max.efficiency ? model : max
    );

    // Generate recommendations
    if (mostExpensive.avgCostPerRequest > 0.01) {
      insights.push({
        type: 'warning',
        title: 'High Cost Model Detected',
        message: `${mostExpensive.name} has high cost per request ($${mostExpensive.avgCostPerRequest.toFixed(4)}). Consider optimizing prompts or switching models.`,
        icon: 'exclamation triangle'
      });
    }

    if (mostEfficient.efficiency > 0) {
      insights.push({
        type: 'success',
        title: 'Most Efficient Model',
        message: `${mostEfficient.name} offers the best token-to-cost ratio. Consider using it for similar tasks.`,
        icon: 'thumbs up'
      });
    }

    // Budget projection
    const totalQuotaToday = summaryData.todayQuota;
    const quotaPerUnit = parseFloat(localStorage.getItem('quota_per_unit') || '500000');
    const dailySpendingInCurrency = totalQuotaToday / quotaPerUnit;
    const monthlyProjection = dailySpendingInCurrency * 30;

    if (monthlyProjection > 100) {
      insights.push({
        type: 'info',
        title: 'Monthly Spending Projection',
        message: `Based on today's usage, monthly spending could reach $${monthlyProjection.toFixed(2)}. Consider setting usage limits.`,
        icon: 'chart line'
      });
    }

    return insights;
  };

  const timeSeriesData = processTimeSeriesData();
  const modelData = processModelData();
  const models = getUniqueModels();

  // 生成随机颜色
  const getRandomColor = (index) => {
    return chartConfig.barColors[index % chartConfig.barColors.length];
  };

  // 添加一个日期格式化函数
  const formatDate = (dateStr) => {
    const date = new Date(dateStr);
    return date.toLocaleDateString('zh-CN', {
      month: 'numeric',
      day: 'numeric',
    });
  };

  // Gradient definitions component
  const GradientDefs = () => (
    <defs>
      <linearGradient id="requestsGradient" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#4318FF" stopOpacity={0.8}/>
        <stop offset="100%" stopColor="#4318FF" stopOpacity={0.1}/>
      </linearGradient>
      <linearGradient id="quotaGradient" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#00B5D8" stopOpacity={0.8}/>
        <stop offset="100%" stopColor="#00B5D8" stopOpacity={0.1}/>
      </linearGradient>
      <linearGradient id="tokensGradient" x1="0" y1="0" x2="0" y2="1">
        <stop offset="0%" stopColor="#FF5E7D" stopOpacity={0.8}/>
        <stop offset="100%" stopColor="#FF5E7D" stopOpacity={0.1}/>
      </linearGradient>
    </defs>
  );

  // Loading skeleton component
  const LoadingSkeleton = ({ height = 140 }) => (
    <div className="loading-skeleton" style={{ height }}>
      <div className="skeleton-content">
        <div className="skeleton-line skeleton-title"></div>
        <div className="skeleton-chart"></div>
      </div>
    </div>
  );

  // 修改所有 XAxis 配置
  const xAxisConfig = {
    dataKey: 'date',
    axisLine: false,
    tickLine: false,
    tick: {
      fontSize: 12,
      fill: 'var(--text-secondary)',
      textAnchor: 'middle', // 文本居中对齐
    },
    tickFormatter: formatDate,
    interval: 0,
    minTickGap: 5,
    padding: { left: 30, right: 30 }, // 增加两侧的内边距，确保首尾标签完整显示
  };

  return (
    <div className='dashboard-container'>
      {/* Controls for root users and date range */}
      <Card fluid style={{ marginBottom: '1rem' }}>
        <Card.Content>
          <div className="controls-header">
            <Card.Header>{t('dashboard.controls.title', 'Dashboard Controls')}</Card.Header>
            <div className="real-time-status">
              <div className="status-indicator">
                <div className="status-dot"></div>
                <span className="status-text">
                  {lastUpdated ?
                    `${t('dashboard.status.updated', 'Updated')}: ${lastUpdated.toLocaleTimeString()}` :
                    t('dashboard.status.loading', 'Loading...')
                  }
                </span>
              </div>
            </div>
          </div>

          {/* User selector for root users */}
          {isRootUser && (
            <Form.Field style={{ marginBottom: '1rem' }}>
              <label>{t('dashboard.user_selector.title', 'Select User')}</label>
              <Dropdown
                placeholder={t('dashboard.user_selector.placeholder', 'Search and select a user to view dashboard')}
                fluid
                selection
                search
                clearable
                value={selectedUserId}
                onChange={handleUserChange}
                options={users.map(user => ({
                  key: user.id,
                  value: user.id === 0 ? 'all' : user.id.toString(),
                  text: user.id === 0 ? user.display_name : `${user.display_name || user.username} (${user.username})`,
                  description: user.id === 0 ? 'View site-wide statistics' : `User ID: ${user.id}`
                }))}
                noResultsMessage={t('dashboard.user_selector.no_results', 'No users found')}
              />
            </Form.Field>
          )}

          {/* Date range preset buttons */}
          <div className="date-presets">
            <Button.Group size="small">
              <Button
                onClick={() => handlePresetDateRange('today')}
                className="preset-button"
              >
                {t('dashboard.presets.today', 'Today')}
              </Button>
              <Button
                onClick={() => handlePresetDateRange('7days')}
                className="preset-button"
              >
                {t('dashboard.presets.7days', 'Last 7 Days')}
              </Button>
              <Button
                onClick={() => handlePresetDateRange('30days')}
                className="preset-button"
              >
                {t('dashboard.presets.30days', 'Last 30 Days')}
              </Button>
            </Button.Group>
          </div>

          {/* Date range selector */}
          <Form>
            <Form.Group widths='equal'>
              <Form.Field>
                <label>{t('dashboard.date_range.from', 'From Date')}</label>
                <input
                  type="date"
                  value={fromDate}
                  min={getMinDate()}
                  max={getMaxDate()}
                  onChange={(e) => handleDateChange('from', e.target.value)}
                  className="modern-date-input"
                />
              </Form.Field>
              <Form.Field>
                <label>{t('dashboard.date_range.to', 'To Date')}</label>
                <input
                  type="date"
                  value={toDate}
                  min={getMinDate()}
                  max={getMaxDate()}
                  onChange={(e) => handleDateChange('to', e.target.value)}
                  className="modern-date-input"
                />
              </Form.Field>
              <Form.Field>
                <label style={{ visibility: 'hidden' }}>Refresh</label>
                <Button
                  primary
                  onClick={handleRefresh}
                  loading={isRefreshing}
                  disabled={isRefreshing}
                  className="refresh-button"
                >
                  <Icon name="refresh" />
                  {t('dashboard.refresh', 'Refresh')}
                </Button>
              </Form.Field>
            </Form.Group>

            {dateError && (
              <Message negative>
                <Message.Header>{t('dashboard.error', 'Error')}</Message.Header>
                <p>{dateError}</p>
              </Message>
            )}

            {!dateError && (
              <div style={{ fontSize: '0.85em', color: '#666', marginTop: '8px' }}>
                {isRootUser ? (
                  t('dashboard.date_range.info_root', 'As root user, you can select up to 1 year of data.')
                ) : (
                  t('dashboard.date_range.info_user', 'Regular users can view up to 7 days of data.')
                )}
              </div>
            )}

          </Form>
        </Card.Content>
      </Card>

      {/* Modern Summary Cards */}
      <Grid columns={4} stackable className='summary-cards-grid'>
        <Grid.Column>
          <Card fluid className='summary-card requests-card'>
            <Card.Content>
              <div className='summary-card-content'>
                <div className='summary-icon-wrapper'>
                  <Icon name='chart line' size='large' />
                </div>
                <div className='summary-stats'>
                  <Statistic size='small'>
                    <Statistic.Value>{renderNumberWithTooltip(summaryData.todayRequests)}</Statistic.Value>
                    <Statistic.Label>{t('dashboard.summary.requests', 'Total Requests')}</Statistic.Label>
                  </Statistic>
                  <div className='trend-indicator'>
                    <Icon
                      name={summaryData.requestTrend >= 0 ? 'arrow up' : 'arrow down'}
                      color={summaryData.requestTrend >= 0 ? 'green' : 'red'}
                      size='small'
                    />
                    <span className={`trend-value ${summaryData.requestTrend >= 0 ? 'positive' : 'negative'}`}>
                      {Math.abs(summaryData.requestTrend).toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='summary-card quota-card'>
            <Card.Content>
              <div className='summary-card-content'>
                <div className='summary-icon-wrapper'>
                  <Icon name='dollar sign' size='large' />
                </div>
                <div className='summary-stats'>
                  <Statistic size='small'>
                    <Statistic.Value>{renderQuota(summaryData.todayQuota, t, 3)}</Statistic.Value>
                    <Statistic.Label>{t('dashboard.summary.quota', 'Total Quota Used')}</Statistic.Label>
                  </Statistic>
                  <div className='trend-indicator'>
                    <Icon
                      name={summaryData.quotaTrend >= 0 ? 'arrow up' : 'arrow down'}
                      color={summaryData.quotaTrend >= 0 ? 'red' : 'green'}
                      size='small'
                    />
                    <span className={`trend-value ${summaryData.quotaTrend >= 0 ? 'negative' : 'positive'}`}>
                      {Math.abs(summaryData.quotaTrend).toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='summary-card tokens-card'>
            <Card.Content>
              <div className='summary-card-content'>
                <div className='summary-icon-wrapper'>
                  <Icon name='code' size='large' />
                </div>
                <div className='summary-stats'>
                  <Statistic size='small'>
                    <Statistic.Value>{renderNumberWithTooltip(summaryData.todayTokens)}</Statistic.Value>
                    <Statistic.Label>{t('dashboard.summary.tokens', 'Total Tokens Used')}</Statistic.Label>
                  </Statistic>
                  <div className='trend-indicator'>
                    <Icon
                      name={summaryData.tokenTrend >= 0 ? 'arrow up' : 'arrow down'}
                      color={summaryData.tokenTrend >= 0 ? 'orange' : 'green'}
                      size='small'
                    />
                    <span className={`trend-value ${summaryData.tokenTrend >= 0 ? 'neutral' : 'positive'}`}>
                      {Math.abs(summaryData.tokenTrend).toFixed(1)}%
                    </span>
                  </div>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='summary-card efficiency-card'>
            <Card.Content>
              <div className='summary-card-content'>
                <div className='summary-icon-wrapper'>
                  <Icon name='lightning' size='large' />
                </div>
                <div className='summary-stats'>
                  <Statistic size='small'>
                    <Statistic.Value>${summaryData.avgCostPerRequest.toFixed(4)}</Statistic.Value>
                    <Statistic.Label>{t('dashboard.summary.avg_cost', 'Avg Cost/Request')}</Statistic.Label>
                  </Statistic>
                  <div className='secondary-metric'>
                    <span className='metric-label'>Avg Tokens:</span>
                    <span className='metric-value'>{renderNumberWithTooltip(Math.round(summaryData.avgTokensPerRequest))}</span>
                  </div>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* Additional Insights Cards */}
      <Grid columns={3} stackable className='insights-cards-grid'>
        <Grid.Column>
          <Card fluid className='insight-card model-insights'>
            <Card.Content>
              <Card.Header>
                <Icon name='brain' />
                {t('dashboard.insights.model_usage', 'Model Usage Insights')}
              </Card.Header>
              <div className='insight-content'>
                <div className='insight-item'>
                  <span className='insight-label'>{t('dashboard.insights.top_model', 'Most Used Model:')}</span>
                  <span className='insight-value highlight'>{summaryData.topModel || 'N/A'}</span>
                </div>
                <div className='insight-item'>
                  <span className='insight-label'>{t('dashboard.insights.total_models', 'Active Models:')}</span>
                  <span className='insight-value'>{summaryData.totalModels}</span>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='insight-card performance-insights'>
            <Card.Content>
              <Card.Header>
                <Icon name='tachometer alternate' />
                {t('dashboard.insights.performance', 'Performance Metrics')}
              </Card.Header>
              <div className='insight-content'>
                <div className='insight-item'>
                  <span className='insight-label'>{t('dashboard.insights.avg_response', 'Avg Response Time:')}</span>
                  <span className='insight-value'>
                    {summaryData.avgResponseTime.toFixed(0)}ms
                  </span>
                </div>
                <div className='insight-item'>
                  <span className='insight-label'>{t('dashboard.insights.success_rate', 'Success Rate:')}</span>
                  <span className='insight-value'>
                    {summaryData.successRate.toFixed(1)}%
                  </span>
                </div>
                <div className='insight-item'>
                  <span className='insight-label'>{t('dashboard.insights.throughput', 'Throughput:')}</span>
                  <span className='insight-value'>
                    {summaryData.throughput.toFixed(1)} req/hr
                  </span>
                </div>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='insight-card usage-patterns'>
            <Card.Content>
              <Card.Header>
                <Icon name='clock outline' />
                {t('dashboard.insights.usage_patterns', 'Usage Patterns')}
              </Card.Header>
              <div className='insight-content'>
                {(() => {
                  const patterns = getUsagePatterns();
                  if (!patterns) return (
                    <div className='insight-item'>
                      <span className='insight-label'>No data available</span>
                    </div>
                  );

                  return (
                    <>
                      <div className='insight-item'>
                        <span className='insight-label'>{t('dashboard.insights.peak_day', 'Peak Day:')}</span>
                        <span className='insight-value'>
                          {new Date(patterns.peakDay).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
                        </span>
                      </div>
                      <div className='insight-item'>
                        <span className='insight-label'>{t('dashboard.insights.daily_avg', 'Daily Average:')}</span>
                        <span className='insight-value'>
                          {patterns.avgDailyRequests.toFixed(0)} req
                        </span>
                      </div>
                      <div className='insight-item'>
                        <span className='insight-label'>{t('dashboard.insights.trend', 'Trend:')}</span>
                        <span className={`insight-value ${
                          patterns.trendDirection === 'increasing' ? 'positive' :
                          patterns.trendDirection === 'decreasing' ? 'negative' : 'neutral'
                        }`}>
                          {patterns.trendDirection === 'increasing' ? '↗ Growing' :
                           patterns.trendDirection === 'decreasing' ? '↘ Declining' : '→ Stable'}
                        </span>
                      </div>
                    </>
                  );
                })()}
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 三个并排的折线图 */}
      <Grid columns={3} stackable className='charts-grid'>
        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.requests.title')}
                {/* <span className='stat-value'>{summaryData.todayRequests}</span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={140}
                  margin={{ left: 10, right: 10 }}
                >
                  <LineChart data={timeSeriesData}>
                    <GradientDefs />
                    <CartesianGrid
                      strokeDasharray='2 2'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                      stroke='var(--border-color)'
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: 'var(--card-bg)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '12px',
                        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
                        color: 'var(--text-primary)',
                        padding: '12px 16px',
                      }}
                      formatter={(value) => [
                        <span style={{ color: chartConfig.colors.requests, fontWeight: '600' }}>
                          {renderNumberForChart(value)}
                        </span>,
                        t('dashboard.charts.requests.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        <div style={{ fontWeight: '600', marginBottom: '4px' }}>
                          {formatDate(label)}
                        </div>
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='requests'
                      stroke={chartConfig.colors.requests}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={{
                        ...chartConfig.lineChart.line.activeDot,
                        fill: chartConfig.colors.requests,
                      }}
                      filter='drop-shadow(0 2px 4px rgba(67, 24, 255, 0.3))'
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.quota.title')}
                {/* <span className='stat-value'>
                  ${summaryData.todayQuota.toFixed(3)}
                </span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={140}
                  margin={{ left: 10, right: 10 }}
                >
                  <LineChart data={timeSeriesData}>
                    <GradientDefs />
                    <CartesianGrid
                      strokeDasharray='2 2'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                      stroke='var(--border-color)'
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: 'var(--card-bg)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '12px',
                        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
                        color: 'var(--text-primary)',
                        padding: '12px 16px',
                      }}
                      formatter={(value) => [
                        <span style={{ color: chartConfig.colors.quota, fontWeight: '600' }}>
                          {renderQuota(value, t, 6)}
                        </span>,
                        t('dashboard.charts.quota.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        <div style={{ fontWeight: '600', marginBottom: '4px' }}>
                          {formatDate(label)}
                        </div>
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='quota'
                      stroke={chartConfig.colors.quota}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={{
                        ...chartConfig.lineChart.line.activeDot,
                        fill: chartConfig.colors.quota,
                      }}
                      filter='drop-shadow(0 2px 4px rgba(0, 181, 216, 0.3))'
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>

        <Grid.Column>
          <Card fluid className='chart-card'>
            <Card.Content>
              <Card.Header>
                {t('dashboard.charts.tokens.title')}
                {/* <span className='stat-value'>{summaryData.todayTokens}</span> */}
              </Card.Header>
              <div className='chart-container'>
                <ResponsiveContainer
                  width='100%'
                  height={140}
                  margin={{ left: 10, right: 10 }}
                >
                  <LineChart data={timeSeriesData}>
                    <GradientDefs />
                    <CartesianGrid
                      strokeDasharray='2 2'
                      vertical={chartConfig.lineChart.grid.vertical}
                      horizontal={chartConfig.lineChart.grid.horizontal}
                      opacity={chartConfig.lineChart.grid.opacity}
                      stroke='var(--border-color)'
                    />
                    <XAxis {...xAxisConfig} />
                    <YAxis hide={true} />
                    <Tooltip
                      contentStyle={{
                        background: 'var(--card-bg)',
                        border: '1px solid var(--border-color)',
                        borderRadius: '12px',
                        boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
                        color: 'var(--text-primary)',
                        padding: '12px 16px',
                      }}
                      formatter={(value) => [
                        <span style={{ color: chartConfig.colors.tokens, fontWeight: '600' }}>
                          {renderNumberForChart(value)}
                        </span>,
                        t('dashboard.charts.tokens.tooltip'),
                      ]}
                      labelFormatter={(label) =>
                        <div style={{ fontWeight: '600', marginBottom: '4px' }}>
                          {formatDate(label)}
                        </div>
                      }
                    />
                    <Line
                      type='monotone'
                      dataKey='tokens'
                      stroke={chartConfig.colors.tokens}
                      strokeWidth={chartConfig.lineChart.line.strokeWidth}
                      dot={chartConfig.lineChart.line.dot}
                      activeDot={{
                        ...chartConfig.lineChart.line.activeDot,
                        fill: chartConfig.colors.tokens,
                      }}
                      filter='drop-shadow(0 2px 4px rgba(255, 94, 125, 0.3))'
                    />
                  </LineChart>
                </ResponsiveContainer>
              </div>
            </Card.Content>
          </Card>
        </Grid.Column>
      </Grid>

      {/* 模型使用统计 */}
      <Card fluid className='chart-card'>
        <Card.Content>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
            <Card.Header>
              {t('dashboard.statistics.title')} - {
                statisticsMetric === 'tokens' ? t('dashboard.statistics.metrics.tokens', 'Tokens') :
                statisticsMetric === 'requests' ? t('dashboard.statistics.metrics.requests', 'Requests') :
                t('dashboard.statistics.metrics.expenses', 'Expenses')
              }
            </Card.Header>
            <Dropdown
              selection
              value={statisticsMetric}
              onChange={handleStatisticsMetricChange}
              options={[
                { key: 'tokens', value: 'tokens', text: t('dashboard.statistics.metrics.tokens', 'Tokens') },
                { key: 'requests', value: 'requests', text: t('dashboard.statistics.metrics.requests', 'Requests') },
                { key: 'expenses', value: 'expenses', text: t('dashboard.statistics.metrics.expenses', 'Expenses') }
              ]}
              style={{ minWidth: '120px' }}
            />
          </div>
          <div className='chart-container'>
            <ResponsiveContainer width='100%' height={350}>
              <BarChart data={modelData} margin={{ top: 20, right: 30, left: 20, bottom: 5 }}>
                <CartesianGrid
                  strokeDasharray='2 2'
                  vertical={false}
                  opacity={0.2}
                  stroke='var(--border-color)'
                />
                <XAxis {...xAxisConfig} />
                <YAxis
                  axisLine={false}
                  tickLine={false}
                  tick={{ fontSize: 12, fill: 'var(--text-secondary)' }}
                  tickFormatter={(value) => {
                    switch (statisticsMetric) {
                      case 'requests':
                        return renderNumberForChart(value);
                      case 'expenses':
                        return `$${value.toFixed(2)}`;
                      case 'tokens':
                      default:
                        return renderNumberForChart(value);
                    }
                  }}
                />
                <Tooltip
                  contentStyle={{
                    background: 'var(--card-bg)',
                    border: '1px solid var(--border-color)',
                    borderRadius: '12px',
                    boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
                    color: 'var(--text-primary)',
                    padding: '12px 16px',
                  }}
                  content={({ active, payload, label }) => {
                    if (active && payload && payload.length) {
                      // Filter out entries with zero values and sort by value in descending order
                      // This ensures the tooltip shows models ordered by their usage for this specific day
                      const filteredAndSortedPayload = payload
                        .filter(entry => entry.value > 0)
                        .sort((a, b) => b.value - a.value);

                      // Format value based on selected metric
                      const formatValue = (value) => {
                        switch (statisticsMetric) {
                          case 'requests':
                            return renderNumberForChart(value);
                          case 'expenses':
                            return `$${value.toFixed(6)}`;
                          case 'tokens':
                          default:
                            return renderNumberForChart(value);
                        }
                      };

                      return (
                        <div style={{
                          background: 'var(--card-bg)',
                          border: '1px solid var(--border-color)',
                          borderRadius: '12px',
                          boxShadow: '0 8px 32px rgba(0, 0, 0, 0.12)',
                          color: 'var(--text-primary)',
                          padding: '12px 16px',
                        }}>
                          <div style={{ fontWeight: '600', marginBottom: '8px' }}>
                            {formatDate(label)}
                          </div>
                          {filteredAndSortedPayload.map((entry, index) => (
                            <div key={index} style={{ marginBottom: '4px' }}>
                              <span style={{
                                display: 'inline-block',
                                width: '12px',
                                height: '12px',
                                backgroundColor: entry.color,
                                borderRadius: '50%',
                                marginRight: '8px'
                              }}></span>
                              <span style={{ fontWeight: '600' }}>
                                {entry.name}: {formatValue(entry.value)}
                              </span>
                            </div>
                          ))}
                        </div>
                      );
                    }
                    return null;
                  }}
                />
                <Legend
                  wrapperStyle={{
                    paddingTop: '24px',
                    fontSize: '14px',
                  }}
                  iconType='circle'
                />
                {models.map((model, index) => (
                  <Bar
                    key={model}
                    dataKey={model}
                    stackId='a'
                    fill={getRandomColor(index)}
                    name={model}
                    radius={index === models.length - 1 ? [4, 4, 0, 0] : [0, 0, 0, 0]}
                  />
                ))}
              </BarChart>
            </ResponsiveContainer>
          </div>
        </Card.Content>
      </Card>

      {/* Model Efficiency Analysis */}
      <Card fluid className='efficiency-analysis-card'>
        <Card.Content>
          <Card.Header>
            <Icon name='chart bar' />
            {t('dashboard.efficiency.title', 'Model Efficiency Analysis')}
          </Card.Header>
          <div className='efficiency-table-container'>
            <div className='efficiency-table'>
              <div className='table-header'>
                <div className='header-cell'>{t('dashboard.efficiency.model', 'Model')}</div>
                <div className='header-cell'>{t('dashboard.efficiency.requests', 'Requests')}</div>
                <div className='header-cell'>{t('dashboard.efficiency.avg_cost', 'Avg Cost')}</div>
                <div className='header-cell'>{t('dashboard.efficiency.avg_tokens', 'Avg Tokens')}</div>
                <div className='header-cell'>{t('dashboard.efficiency.efficiency', 'Efficiency')}</div>
              </div>
              <div className='table-body'>
                {getModelEfficiencyData().slice(0, 5).map((model, index) => (
                  <div key={model.name} className='table-row'>
                    <div className='table-cell model-name'>
                      <div className='model-rank'>{index + 1}</div>
                      <span>{model.name}</span>
                    </div>
                    <div className='table-cell'>{renderNumberWithTooltip(model.requests)}</div>
                    <div className='table-cell'>${model.avgCostPerRequest.toFixed(4)}</div>
                    <div className='table-cell'>{model.avgTokensPerRequest.toFixed(0)}</div>
                    <div className='table-cell'>
                      <div className='efficiency-bar'>
                        <div
                          className='efficiency-fill'
                          style={{
                            width: `${Math.min(100, (model.efficiency / Math.max(...getModelEfficiencyData().map(m => m.efficiency))) * 100)}%`
                          }}
                        ></div>
                        <span className='efficiency-text'>{model.efficiency.toFixed(0)}</span>
                      </div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          </div>
        </Card.Content>
      </Card>

      {/* Cost Optimization Recommendations */}
      {getCostOptimizationInsights().length > 0 && (
        <Card fluid className='optimization-recommendations-card'>
          <Card.Content>
            <Card.Header>
              <Icon name='lightbulb outline' />
              {t('dashboard.optimization.title', 'Cost Optimization Recommendations')}
            </Card.Header>
            <div className='recommendations-container'>
              {getCostOptimizationInsights().map((insight, index) => (
                <div key={index} className={`recommendation-item ${insight.type}`}>
                  <div className='recommendation-icon'>
                    <Icon name={insight.icon} />
                  </div>
                  <div className='recommendation-content'>
                    <div className='recommendation-title'>{insight.title}</div>
                    <div className='recommendation-message'>{insight.message}</div>
                  </div>
                </div>
              ))}
            </div>
          </Card.Content>
        </Card>
      )}
    </div>
  );
};

export default Dashboard;
