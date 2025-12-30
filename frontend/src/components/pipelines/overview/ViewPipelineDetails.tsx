import { Boxes } from "lucide-react";
import { useEffect, useState, useCallback, useRef } from "react";
import pipelineServices from "@/services/pipeline";
import { PipelineOverviewInterface } from "@/types/pipeline.types";
import PipelinYAML from "./YamlViewer";
import PipelineOverview from "./PipelineOverview";
import { useGlobalSnackbar } from "@/context/useGlobalSnackbar";
import DeletePipelineDialog from "./DeletePipelineDialog";
import CloseIcon from "@mui/icons-material/Close";
import { IconButton, Button, Drawer, Typography, Box, Tabs, Tab } from "@mui/material";
import { useNavigate } from "react-router-dom";

interface Props {
  pipelineId: string;
  open: boolean;
  onClose: () => void;
}

const MIN_WIDTH = 800;
const MAX_WIDTH = 1200;

const ViewPipelineDetails = ({ pipelineId, open, onClose }: Props) => {
	const [width, setWidth] = useState(900);
	const [pipelineOverviewData, setPipelineOverviewData] =
		useState<PipelineOverviewInterface>();
	const [tabs, setTabs] = useState("overview");
	const [isDeleteOpen, setIsDeleteOpen] = useState(false);
	const navigate = useNavigate();
	const { showSnackbar } = useGlobalSnackbar();
	const resizingRef = useRef(false);

	const handleGetPipelineOverview = useCallback(async () => {
		try {
			const res = await pipelineServices.getPipelineOverviewById(pipelineId);
			setPipelineOverviewData(res);
		} catch {
			showSnackbar("Failed to fetch pipeline overview", "error");
		}
	}, [pipelineId, showSnackbar]);

	useEffect(() => {
		handleGetPipelineOverview();
	}, [handleGetPipelineOverview]);


	const startResize = () => {
		resizingRef.current = true;
		document.addEventListener("mousemove", resize);
		document.addEventListener("mouseup", stopResize);
	};

	const resize = (e: MouseEvent) => {
		if (!resizingRef.current) return;
		const newWidth = window.innerWidth - e.clientX;
		if (newWidth >= MIN_WIDTH && newWidth <= MAX_WIDTH) {
			setWidth(newWidth);
		}
	};

	const stopResize = () => {
		resizingRef.current = false;
		document.removeEventListener("mousemove", resize);
		document.removeEventListener("mouseup", stopResize);
	};

	const handleAttemptClose = () => {
		onClose();
	};
	
	return (
		<>
			<Drawer
				anchor="right"
				open={open}
				onClose={handleAttemptClose}
				PaperProps={{
				sx: {
					width,
					display: "flex",
					flexDirection: "row",
				},
				}}
			>
				{/* Drag Handle */}
				<div
					onMouseDown={startResize}
					style={{
						width: 10,
						display: "flex",
						alignItems: "center",
						justifyContent: "center",
						cursor: "col-resize",
					}}
				>
					<div
						style={{
						width: 4,
						height: 48,
						borderRadius: 2,
						backgroundColor: "#93c5fd",
						}}
					/>
				</div>
				{/* Content */}
				<div className="flex flex-col w-full overflow-hidden">
				{/* Header */}
					<div className="relative flex justify-between items-center pr-10 py-2 border-b">
						<div className="flex gap-2 items-center">
						<Boxes size={28} />
						<Typography variant="h6">{pipelineOverviewData?.name}</Typography>
						</div>
						<div className="flex gap-2">
						<Button
							variant="contained"
							color="primary"
							onClick={() => {
								navigate(`/pipelines/${pipelineId}/edit`, {
									state: { pipelineName: pipelineOverviewData?.name },
								});
							}}
							sx={{ textTransform: "none" }}
						>
							View / Edit Pipeline
						</Button>
						<DeletePipelineDialog
							isOpen={isDeleteOpen}
							setIsOpen={setIsDeleteOpen}
							pipelineOverview={pipelineOverviewData}
						/>
						</div>
						<IconButton
							size="small"
							onClick={handleAttemptClose}
							sx={{
								position: "absolute",
								top: 3,
								right: 3,
								width: 28,
								height: 28,
								borderRadius: "7px",
								color: "grey.600",
								padding: 0,
								"&:hover": {
								backgroundColor: "error.main",
								color: "#fff",
								},
							}}
						>
							<CloseIcon fontSize="small" />
						</IconButton>
					</div>
					{/* Tabs */}
					<Box sx={{ borderBottom: 1, borderColor: "divider" }}>
						<Tabs
							value={tabs}
							onChange={(_, newValue) => setTabs(newValue)}
							textColor="primary"
							indicatorColor="primary"
						>
							<Tab label="Overview" value="overview" />
							<Tab label="YAML" value="yaml" />
						</Tabs>
					</Box>
					{/* Body */}
					<div className="flex-1 overflow-auto p-4">
						{tabs === "overview" && <PipelineOverview pipelineId={pipelineId} />}
						{tabs === "yaml" && (
						<PipelinYAML jsonforms={pipelineOverviewData?.config} />
						)}
					</div>
				</div>
			</Drawer>
		</>
	);
};

export default ViewPipelineDetails;
