import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	Typography,
	FormControl,
	InputLabel,
	Select,
	MenuItem,
	Box,
	CircularProgress,
} from "@mui/material";
import { useEffect, useState } from "react";
import agentServices from "@/services/agent";
import pipelineServices from "@/services/pipeline";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";

interface Pipeline {
	id: string;
	name: string;
}

interface AttachAgentDialogProps {
	open: boolean;
	agentId: number;
	onClose: () => void;
	onSuccess?: () => void;
}

const AttachAgentDialog = ({
	open,
	agentId,
	onClose,
	onSuccess,
}: AttachAgentDialogProps) => {
	const { showSnackbar } = useGlobalSnackbar();

	const [pipelines, setPipelines] = useState<Pipeline[]>([]);
	const [selectedPipeline, setSelectedPipeline] = useState<string>("");
	const [loading, setLoading] = useState(false);
	const [fetching, setFetching] = useState(false);

	// ---------------------------
	// Fetch pipelines
	// ---------------------------
	useEffect(() => {
		if (!open) return;

		const fetchPipelines = async () => {
			try {
				setFetching(true);
				const data = await pipelineServices.getAllPipelines();
				setPipelines(data || []);
			} catch {
				showSnackbar("Failed to fetch pipelines", "error");
			} finally {
				setFetching(false);
			}
		};

		fetchPipelines();
	}, [open, showSnackbar]);

	const handleAttach = async () => {
		if (!agentId || !selectedPipeline) return;

		try {
			setLoading(true);
			await agentServices.attachAgentToPipeline(
				Number(selectedPipeline),
				agentId
			);
			showSnackbar("Agent attached successfully", "success");
			onSuccess?.();
			onClose();
		} catch (error) {
			showSnackbar(
				error instanceof Error
					? error.message
					: "Failed to attach agent",
				"error"
			);
		} finally {
			setLoading(false);
		}
	};

	const noPipelines = !fetching && pipelines.length === 0;

	return (
		<Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
			<DialogTitle>Attach Agent</DialogTitle>

			<DialogContent>
				<Typography variant="body2" color="text.secondary" sx={{ mb: 2 }}>
					Select a pipeline to attach this agent to.
				</Typography>

				{fetching ? (
					<Box className="flex justify-center py-4">
						<CircularProgress size={20} />
					</Box>
				) : noPipelines ? (
					<Box
						sx={{
							p: 2,
							border: "1px dashed",
							borderColor: "grey.300",
							borderRadius: 1,
							backgroundColor: "grey.50",
						}}
					>
						<Typography
							variant="body2"
							color="text.secondary"
							align="center"
						>
							No pipelines available.
							<br />
							Create a pipeline to attach this agent.
						</Typography>
					</Box>
				) : (
					<FormControl fullWidth size="small">
						<InputLabel id="pipeline-select-label">
							Pipeline
						</InputLabel>
						<Select
							labelId="pipeline-select-label"
							label="Pipeline"
							value={selectedPipeline}
							onChange={e =>
								setSelectedPipeline(e.target.value)
							}
						>
							{pipelines.map(pipeline => (
								<MenuItem
									key={pipeline.id}
									value={pipeline.id}
								>
									{pipeline.name}
								</MenuItem>
							))}
						</Select>
					</FormControl>
				)}
			</DialogContent>

			<DialogActions>
				<Button onClick={onClose} disabled={loading}>
					Cancel
				</Button>
				<Button
					variant="contained"
					onClick={handleAttach}
					disabled={
						loading || noPipelines || !selectedPipeline
					}
				>
					{loading ? "Attaching..." : "Attach"}
				</Button>
			</DialogActions>
		</Dialog>
	);
};

export default AttachAgentDialog;
