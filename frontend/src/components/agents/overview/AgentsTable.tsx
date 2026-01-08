import {
	Table,
	TableBody,
	TableCell,
	TableContainer,
	TableHead,
	TableRow,
	Paper,
} from "@mui/material";
import { useEffect, useState } from "react";
import agentServices from "@/services/agent";
import ViewAgentDetails from "./ViewAgentDetails";
import { Agent } from "@/types/agent.types";

const AgentsTable = () => {
	const [agents, setAgents] = useState<Agent[]>([]);
	const [agentId, setAgentId] = useState<number | null>(null);
	const [drawerOpen, setDrawerOpen] = useState(false);

	const handleGetAgents = async () => {
		const res = await agentServices.getAllAgents();
		setAgents(res);
	};

	useEffect(() => {
		handleGetAgents();
	}, []);

	const handleRowClick = (id: number) => {
		setAgentId(id);
		setDrawerOpen(true);
	};

	const handleCloseDrawer = () => {
		setDrawerOpen(false);
		setAgentId(null);
	};

	if (!agents || agents.length === 0) {
		return (
			<div className="flex flex-col gap-2 justify-center items-center">
				<p className="font-bold text-xl mt-[6rem]">No Agents Found</p>
				<p className="text-gray-700">
					Agents process and route data within your system.
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
							<TableCell>ID</TableCell>
							<TableCell>Name</TableCell>
							<TableCell>Type</TableCell>
							<TableCell>Version</TableCell>
							<TableCell>Status</TableCell>
						</TableRow>
					</TableHead>

					<TableBody>
						{agents.map(agent => (
							<TableRow
								key={agent.id}
								hover
								sx={{ cursor: "pointer" }}
								onClick={() => handleRowClick(agent.id)}
							>
								<TableCell>{agent.id}</TableCell>
								<TableCell>{agent.name}</TableCell>
								<TableCell>{agent.type}</TableCell>
								<TableCell>{agent.version}</TableCell>
								<TableCell
									sx={{
										fontWeight: 500,
										color:
											agent.status === "connected"
												? "success.main"
												: agent.status === "disconnected"
												? "error.main"
												: "text.secondary",
									}}
								>
									{agent.status}
								</TableCell>
							</TableRow>
						))}
					</TableBody>
				</Table>
			</TableContainer>

			{agentId !== null && (
				<ViewAgentDetails
					agentId={agentId}
					open={drawerOpen}
					onClose={handleCloseDrawer}
				/>
			)}
		</>
	);
};

export default AgentsTable;
