import {
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	Paper,
} from "@mui/material";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { usePipelineOverview } from "@/context/usePipelineDetailContext";
import pipelineServices from "@/services/pipeline";
import { useEffect, useState, useCallback } from "react";
import ViewPipelineDetails from "./ViewPipelineDetails";

interface pipeline {
	id: string;
	name: string;
	agents: number;
	incoming_bytes: number;
	outgoing_bytes: number;
	updatedAt: number;
}

const formatTimestamp = (timestamp: number) => {
	return new Date(timestamp * 1000)
		.toLocaleString("en-GB", {
			day: "2-digit",
			month: "2-digit",
			year: "numeric",
			hour: "2-digit",
			minute: "2-digit",
			second: "2-digit",
			hour12: false,
		})
		.replace(",", "");
};

const PipelineTable = () => {
	const [pipelines, setPipelines] = useState<pipeline[]>([]);
	const [pipelineId, setPipelineId] = useState<string>("");
	const [drawerOpen, setDrawerOpen] = useState(false);

	const { setPipelineOverview } = usePipelineOverview();
	const { resetGraph } = useGraphFlow();

	const handleGetPipelines = async () => {
		const res = await pipelineServices.getAllPipelines();
		setPipelines(res);
	};

	const handleGetPipeline = useCallback(async () => {
		const res = await pipelineServices.getPipelineById(pipelineId);
		setPipelineOverview(res);
	}, [pipelineId, setPipelineOverview]);

	useEffect(() => {
		handleGetPipelines();
	}, []);

	useEffect(() => {
		if (pipelineId) {
			handleGetPipeline();
		}
	}, [pipelineId, handleGetPipeline]);

	const handleRowClick = (id: string) => {
		setPipelineId(id);
		setDrawerOpen(true);
	};

	const handleCloseDrawer = () => {
		setDrawerOpen(false);
		setPipelineId("");
		resetGraph();
		handleGetPipelines();
	};

	if (!pipelines || pipelines.length === 0) {
		return (
			<div className="flex flex-col gap-2 justify-center items-center">
				<p className="font-bold text-xl mt-[6rem]">Get started</p>
				<p className="text-gray-700">Create Your First Pipeline</p>
				<p className="text-gray-700">
					Pipelines collect data from the sources in the pipeline and route them to desired destination.
				</p>
			</div>
		);
	}

	return (
		<>
			<TableContainer component={Paper} variant="outlined">
				<Table>
					<TableHead>
						<TableRow sx={{ backgroundColor: "#f5f5f5" }}>
						<TableCell>Name</TableCell>
						<TableCell>Incoming bytes</TableCell>
						<TableCell>Outgoing bytes</TableCell>
						<TableCell>Updated at</TableCell>
						</TableRow>
					</TableHead>

					<TableBody>
						{pipelines.map(pipeline => (
						<TableRow
							key={pipeline.id}
							hover
							sx={{ cursor: "pointer" }}
							onClick={() => handleRowClick(pipeline.id)}
						>
							<TableCell>{pipeline.name}</TableCell>
							<TableCell>{pipeline.incoming_bytes}</TableCell>
							<TableCell>{pipeline.outgoing_bytes}</TableCell>
							<TableCell>{formatTimestamp(pipeline.updatedAt)}</TableCell>
						</TableRow>
						))}
					</TableBody>
				</Table>
			</TableContainer>

			<ViewPipelineDetails
				pipelineId={pipelineId}
				open={drawerOpen}
				onClose={handleCloseDrawer}
			/>
		</>
	);
};

export default PipelineTable;
