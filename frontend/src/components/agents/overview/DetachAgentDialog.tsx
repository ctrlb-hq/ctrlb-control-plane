import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogActions,
	Button,
	Typography,
	Box,
} from "@mui/material";
import agentServices from "@/services/agent";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import { useState } from "react";

interface DetachAgentDialogProps {
	open: boolean;
	agentId: number;
	pipelineId: number;
	onClose: () => void;
	onSuccess?: () => void;
}

const DetachAgentDialog = ({
	open,
	agentId,
	pipelineId,
	onClose,
	onSuccess,
}: DetachAgentDialogProps) => {
	const { showSnackbar } = useGlobalSnackbar();
	const [loading, setLoading] = useState(false);

	const handleDetach = async () => {
		if (!agentId || !pipelineId) return;

		try {
			setLoading(true);
			await agentServices.detachAgentFromPipeline(pipelineId, agentId);
			showSnackbar("Agent detached successfully", "success");
			onSuccess?.();
			onClose();
		} catch (error) {
			showSnackbar(
				error instanceof Error
					? error.message
					: "Failed to detach agent",
				"error"
			);
		} finally {
			setLoading(false);
		}
	};

	return (
		<Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
			<DialogTitle>Detach Agent</DialogTitle>

			<DialogContent>
				<Box sx={{ mb: 2 }}>
					<Typography variant="body2" color="text.secondary">
						This agent is currently attached to:
					</Typography>

					<Typography
						variant="body1"
						fontWeight={600}
						sx={{ mt: 0.5 }}
					>
						Pipeline ID: {pipelineId}
					</Typography>
				</Box>

				<Typography variant="body2" color="text.secondary">
					Detaching this agent will remove it from the pipeline and may
					stop data collection. Are you sure you want to proceed?
				</Typography>
			</DialogContent>

			<DialogActions>
				<Button onClick={onClose} disabled={loading}>
					Cancel
				</Button>
				<Button
					variant="contained"
					color="error"
					onClick={handleDetach}
					disabled={loading}
				>
					{loading ? "Detaching..." : "Detach"}
				</Button>
			</DialogActions>
		</Dialog>
	);
};

export default DetachAgentDialog;
