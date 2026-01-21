import { 
  Card, 
  CardContent, 
  Typography, 
  Box 
} from '@mui/material';
import {
  AreaChart,
  Area,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  TooltipProps
} from 'recharts';

export interface HealthMetricPoint {
  timestamp: number; 
  [key: string]: number;
}

interface HealthChartProps {
  data: HealthMetricPoint[];
  name: string;
  y_axis_data_key?: string;
  chart_color?: string;
  yAxisLabel?: string;
}

export function HealthChart({
  data,
  name,
  y_axis_data_key = "value",
  chart_color = "#ff9800",
  yAxisLabel,
}: HealthChartProps) {

  const gradientId = `gradient-${name.replace(/\s+/g, '-').toLowerCase()}`;

  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp * 1000);
    const hours = date.getHours().toString().padStart(2, "0");
    const minutes = date.getMinutes().toString().padStart(2, "0");
    const seconds = date.getSeconds().toString().padStart(2, "0");
    return `${hours}:${minutes}:${seconds}`;
  };

  	const CustomTooltip = ({
		active,
		payload,
	}: TooltipProps<number, string>) => {
		if (!active || !payload || payload.length === 0) {
			return null;
		}

		const point = payload[0].payload as HealthMetricPoint;

		return (
			<Box
			sx={{
				backgroundColor: 'rgba(255, 255, 255, 0.95)',
				border: '1px solid #ccc',
				borderRadius: 1,
				padding: 1.5,
				boxShadow: 2,
			}}
			>
			<Typography variant="body2" sx={{ fontWeight: 500 }}>
				Time: {formatTimestamp(point.timestamp)}
			</Typography>
			<Typography variant="body2" color="primary">
				{yAxisLabel ?? 'Value'}: {payload[0].value}
			</Typography>
			</Box>
		);
	};


	return (
		<Card sx={{ width: '100%', boxShadow: 3 }}>
			<CardContent>
				<p className="text-md">
					{name}
				</p>
				<Box sx={{ width: '100%', height: 300, mt: 1 }}>
				<ResponsiveContainer width="100%" height="100%">
					<AreaChart
					data={data}
					margin={{ top: 10, right: 20, left: 0, bottom: 0 }}
					>
					<defs>
						<linearGradient id={gradientId} x1="0" y1="0" x2="0" y2="1">
						<stop offset="5%" stopColor={chart_color} stopOpacity={0.8} />
						<stop offset="95%" stopColor={chart_color} stopOpacity={0.1} />
						</linearGradient>
					</defs>
					<CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
					<XAxis
						dataKey="timestamp"
						tickFormatter={formatTimestamp}
						tick={{ fontSize: 10 }}
						stroke="#666"
					/>
					<YAxis
						label={
						yAxisLabel
							? { value: yAxisLabel, angle: -90, position: 'insideLeft', fontSize: 12 }
							: undefined
						}
						tick={{ fontSize: 10 }}
						stroke="#666"
					/>
					<Tooltip content={<CustomTooltip />} />
					<Area
						type="monotone"
						dataKey={y_axis_data_key}
						stroke={chart_color}
						fill={`url(#${gradientId})`}
						fillOpacity={0.6}
						strokeWidth={2}
					/>
					</AreaChart>
				</ResponsiveContainer>
				</Box>
			</CardContent>
		</Card>
	);
}