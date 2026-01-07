import {
	Dialog,
	DialogTitle,
	DialogContent,
	DialogContentText,
	DialogActions,
	Button,
	Alert,
	Typography,
} from "@mui/material";
import WarningAmberIcon from "@mui/icons-material/WarningAmber";
import pipelineServices from "@/services/pipeline";
import { PipelineOverviewInterface } from "@/types/pipeline.types";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import { useGraphFlow } from "@/context/useGraphFlowContext";

interface Props {
	open: boolean;
	onClose: () => void;
	pipelineOverview?: PipelineOverviewInterface;
}

const DeletePipelineDialog = ({ open, onClose, pipelineOverview }: Props) => {
	const { showSnackbar } = useGlobalSnackbar();
	const { resetGraph } = useGraphFlow();
  
	const handleDeletePipeline = async () => { 
		try {
			if (pipelineOverview?.id) {
				await pipelineServices.deletePipelineById(pipelineOverview.id);
			}
			showSnackbar("Pipeline deleted successfully", "success");
			setIsOpen(false);
			resetGraph();
			onClose();
			window.location.reload();
		} catch (error) {
			console.error("Error deleting pipeline or collector:", error);
			showSnackbar("Failed to delete pipeline or collector", "error");
		}
	};

	return (
		<Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
			<DialogTitle sx={{ color: "error.main", fontWeight: 600 }}>
				Delete Pipeline
			</DialogTitle>

			<DialogContent>
				<DialogContentText sx={{ mb: 2 }}>
					Are you sure you want to permanently delete this pipeline?
					This action <strong>cannot be undone</strong>.
				</DialogContentText>

				<Typography variant="body2">
					<strong>Pipeline ID:</strong> {pipelineOverview?.id}
				</Typography>
				<Typography variant="body2" sx={{ mb: 2 }}>
					<strong>Pipeline Name:</strong> {pipelineOverview?.name}
				</Typography>

				<Alert
					severity="warning"
					icon={<WarningAmberIcon />}
				>
					Deleting this pipeline will stop the collector on the associated
					VM / Kubernetes node.
				</Alert>
			</DialogContent>

			<DialogActions>
				<Button onClick={onClose} variant="outlined">
					Cancel
				</Button>
				<Button
					onClick={handleDeletePipeline}
					variant="contained"
					color="error"
				>
					Delete
				</Button>
			</DialogActions>
		</Dialog>
	);
};

export default DeletePipelineDialog;
