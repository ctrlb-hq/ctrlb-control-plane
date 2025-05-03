import { Area, AreaChart, CartesianGrid, XAxis, YAxis } from "recharts"
import {
    Card,
    CardContent,
    CardHeader,
    CardTitle,
} from "@/components/ui/card"
import {
    ChartConfig,
    ChartContainer,
    ChartTooltip,
    ChartTooltipContent,
} from "@/components/ui/chart"

const chartConfig = {
    desktop: {
        label: "value",
        color: "orange",
    },
} satisfies ChartConfig

export function HealthChart({ data,name,y_axis_data_key, chart_color }: { data: any, name:string, y_axis_data_key?:string, chart_color?:string }) {
    const formatTimestamp = (timestamp: string) => {
        const date = new Date(parseInt(timestamp, 10)*1000)
        const hours = date.getHours().toString().padStart(2, '0')
        const minutes = date.getMinutes().toString().padStart(2, '0')
        const seconds = date.getSeconds().toString().padStart(2, '0')
        return `${hours}:${minutes}:${seconds}`
    }

    return (
        <Card>
            <CardHeader>
                <CardTitle>{name}</CardTitle>
            </CardHeader>
            <CardContent>
                <ChartContainer config={chartConfig}>
                    <AreaChart
                        accessibilityLayer
                        data={data}
                        margin={{
                            left: 12,
                            right: 12,
                        }}
                    >
                        <CartesianGrid vertical={false} />
                        <XAxis
                            dataKey="timestamp"
                            tickLine={false}
                            axisLine={false}
                            tickMargin={8}
                            tickFormatter={formatTimestamp}
                        />
                        <YAxis dataKey={y_axis_data_key || "value"} />
                        <ChartTooltip
                            cursor={false}
                            content={<ChartTooltipContent indicator="line" />}
                        />
                        <Area
                            dataKey={y_axis_data_key || "value"}
                            type="monotone"
                            fill={chart_color || "orange"}
                            fillOpacity={0.4}
                            stroke={chart_color || "orange"}
                        />
                    </AreaChart>
                </ChartContainer>
            </CardContent>
        </Card>
    )
}