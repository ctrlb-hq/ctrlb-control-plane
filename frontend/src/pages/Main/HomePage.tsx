import { useState } from "react";
import { useNavigate } from "react-router-dom";
import authService from "@/services/auth";
import PipelineTable from "@/components/pipelines/overview/PipelineTable";
import AgentsTable from "@/components/agents/overview/AgentsTable";
import { ROUTES } from "@/constants";
import AddPipelineSheet from "@/components/pipelines/create/AddPipelineSheet";
import { Button } from "@mui/material";
import { ArrowLeftRight, Bot, Plus } from "lucide-react";

type TabType = "pipelines" | "agents";

export function HomePage() {
	const navigate = useNavigate();
	const [isAddPipelineOpen, setIsAddPipelineOpen] = useState(false);
	const [activeTab, setActiveTab] = useState<TabType>("pipelines");

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
					{/* Tabs */}
					<div className="flex gap-2 border-b">
						<button
							onClick={() => setActiveTab("pipelines")}
							className={`px-4 py-2 flex items-center gap-2 border-b-2 ${
								activeTab === "pipelines"
									? "border-primary font-medium"
									: "border-transparent text-gray-500"
							}`}
						>
							<ArrowLeftRight size={18} />
							Pipelines
						</button>

						<button
							onClick={() => setActiveTab("agents")}
							className={`px-4 py-2 flex items-center gap-2 border-b-2 ${
								activeTab === "agents"
									? "border-primary font-medium"
									: "border-transparent text-gray-500"
							}`}
						>
							<Bot size={18} />
							Agents
						</button>
					</div>

					{/* Actions */}
					<div className="flex gap-2 items-center">
						{activeTab === "pipelines" && (
							<Button
								variant="contained"
								startIcon={<Plus size={16} />}
								onClick={() => setIsAddPipelineOpen(true)}
								color="primary"
							>
								Add New Pipeline
							</Button>
						)}

						<Button
							variant="contained"
							onClick={handleLogout}
							color="error"
						>
							Logout
						</Button>
					</div>
				</div>

				{/* Content */}
				<div className="p-4">
					{activeTab === "pipelines" ? (
						<PipelineTable />
					) : (
						<AgentsTable />
					)}
				</div>
			</div>

			{/* Controlled Sheet (Pipelines only) */}
			<AddPipelineSheet
				isOpen={isAddPipelineOpen}
				setIsOpen={setIsAddPipelineOpen}
			/>
		</div>
	);
}

export default HomePage;
