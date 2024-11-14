import { Typography } from "@material-tailwind/react";

export function EmptyPipelineMessage() {
    return (
        <div className="flex items-center justify-center h-48 text-gray-500">
            <Typography
                className="text-lg"
                placeholder={undefined}
                onPointerEnterCapture={undefined}
                onPointerLeaveCapture={undefined}
            >
                No data available in Pipelines
            </Typography>
        </div>
    );
}