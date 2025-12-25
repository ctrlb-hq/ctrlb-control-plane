import { useState } from "react";
import { useNavigate } from "react-router-dom";
import authService from "../../services/auth";
import PipelineTable from "@/components/pipelines/overview/PipelineTable";
import { ROUTES } from "../../constants";
import AddPipelineSheet from "@/components/pipelines/create/AddPipelineSheet";
import { Button } from "@mui/material";
import { ArrowLeftRight, Plus } from "lucide-react";

export function HomePage() {
	const navigate = useNavigate();
	const [isAddPipelineOpen, setIsAddPipelineOpen] = useState(false);

	const handleLogout = async () => {
		try {
			await authService.logout();
			navigate(ROUTES.LOGIN, { replace: true });
		} catch {
			localStorage.clear();
			navigate(ROUTES.LOGIN, { replace: true });
		}
	};

	return (
		<div className="w-full h-full">
			<div className="p-4">
				{/* Header */}
				<div className="flex justify-between items-center">
					<div className="flex gap-2 border-b">
						<button className="px-4 py-2 flex items-center gap-2">
							<ArrowLeftRight size={18} />
							Pipelines
						</button>
					</div>

					<div className="flex gap-2 items-center">
						<Button
							variant="contained"
							startIcon={<Plus size={16} />}
							onClick={() => setIsAddPipelineOpen(true)}
							sx={{
								backgroundColor: "#3b82f6",
								"&:hover": {
									backgroundColor: "#2563eb", 
								},
								textTransform: "none",
								boxShadow: "none",
							}}
						>
							Add New Pipeline
						</Button>
						<Button
							variant="contained"
							onClick={handleLogout}
							sx={{
								backgroundColor: "#ef4444",
								"&:hover": {
									backgroundColor: "#dc2626",
								},
								textTransform: "none",
								boxShadow: "none",
							}}
						>
							Logout
						</Button>
					</div>
				</div>

				{/* Table */}
				<div className="p-4">
					<PipelineTable />
				</div>
			</div>

			{/* Controlled Sheet */}
			<AddPipelineSheet
				isOpen={isAddPipelineOpen}
				setIsOpen={setIsAddPipelineOpen}
			/>
		</div>
	);
}

export default HomePage;
