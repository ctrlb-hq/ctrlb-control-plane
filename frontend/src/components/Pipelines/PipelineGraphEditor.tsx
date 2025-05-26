import { Background, Controls, MiniMap, Panel, ReactFlow } from "reactflow";
import { Button } from "@/components/ui/button";
import { Switch } from "@/components/ui/switch";
import { Label } from "@/components/ui/label";
import {
	Sheet,
	SheetTrigger,
	SheetContent,
	SheetTitle,
	SheetDescription,
	SheetClose,
} from "@/components/ui/sheet";
import { Trash2 } from "lucide-react";

import DestinationDropdownOptions from "./DropdownOptions/DestinationDropdownOptions";
import ProcessorDropdownOptions from "./DropdownOptions/ProcessorDropdownOptions";
import SourceDropdownOptions from "./DropdownOptions/SourceDropdownOptions";

interface ChangeLogEntry {
	type: string;
	name: string;
	status: string;
}

interface Props {
	pipelineOverview: { name: string } | null | undefined;
	isEditMode: boolean;
	setIsEditMode: (val: boolean) => void;
	changesLog: ChangeLogEntry[];
	handleDeployChanges: () => void;
	reactFlowWrapper: React.RefObject<HTMLDivElement>;
	nodeValue: any[];
	edgeValue: any[];
	updateNodes: any;
	updateEdges: any;
	onConnect?: any;
	nodeTypes: any;
	setReactFlowInstance: any;
	onEdgeClick: any;
	onPaneClick: any;
	selectedEdge: any;
	edgePopoverPosition: { x: number; y: number };
	handleDeleteEdge: () => void;
}

const PipelineGraphEditor = ({
	pipelineOverview,
	isEditMode,
	setIsEditMode,
	changesLog,
	handleDeployChanges,
	reactFlowWrapper,
	nodeValue,
	edgeValue,
	updateNodes,
	updateEdges,
	onConnect,
	nodeTypes,
	setReactFlowInstance,
	onEdgeClick,
	onPaneClick,
	selectedEdge,
	edgePopoverPosition,
	handleDeleteEdge,
}: Props) => {
	return (
		<>
			{/* Header */}
			<div className="flex justify-between items-center p-4 border-b">
				<div className="text-xl font-medium">{pipelineOverview?.name}</div>
				<div className="flex items-center mr-6">
					<div className="mx-4 flex items-center space-x-2">
						<Switch id="edit-mode" checked={isEditMode} onCheckedChange={setIsEditMode} />
						<Label htmlFor="edit-mode">Edit Mode</Label>
					</div>

					{/* Review Changes Panel */}
					<Sheet>
						<SheetTrigger asChild>
							<Button className="rounded-md px-6" disabled={!isEditMode}>
								Review
							</Button>
						</SheetTrigger>
						<SheetContent className="w-[30rem]">
							<SheetTitle>Pending Changes</SheetTitle>
							<SheetDescription>
								<div className="flex flex-col gap-6 mt-4 overflow-auto h-[40rem]">
									{changesLog.map((change, index) => (
										<div key={index} className="flex justify-between items-center">
											<div className="flex flex-col">
												<p className="text-lg">{change.type}</p>
												<p className="text-lg text-gray-800">{change.name}</p>
											</div>
											<p
												className={`${
													change.status === "added"
														? "text-green-500"
														: change.status === "deleted"
															? "text-red-500"
															: "text-gray-600"
												} text-lg`}>
												[{change.status}]
											</p>
										</div>
									))}
								</div>
							</SheetDescription>
							<SheetClose className="flex justify-end mt-4 w-full">
								<Button onClick={handleDeployChanges} className="bg-blue-500">
									Deploy Changes
								</Button>
							</SheetClose>
						</SheetContent>
					</Sheet>
				</div>
			</div>

			{/* Graph */}
			<div style={{ height: "77vh", backgroundColor: "#f9f9f9" }} ref={reactFlowWrapper}>
				<ReactFlow
					nodes={nodeValue}
					edges={edgeValue}
					onNodesChange={updateNodes}
					onEdgesChange={updateEdges}
					onConnect={isEditMode ? onConnect : undefined}
					nodeTypes={nodeTypes}
					onInit={setReactFlowInstance}
					onEdgeClick={onEdgeClick}
					onPaneClick={onPaneClick}
					nodesDraggable={isEditMode}
					nodesConnectable={isEditMode}
					elementsSelectable={isEditMode}
					fitView>
					<Background />
					<Controls />
					<MiniMap />
					{selectedEdge && isEditMode && (
						<Panel
							position="top-left"
							style={{
								position: "absolute",
								left: edgePopoverPosition.x,
								top: edgePopoverPosition.y,
								transform: "translate(-50%, -50%)",
								background: "white",
								padding: "8px",
								borderRadius: "4px",
								boxShadow: "0 2px 4px rgba(0,0,0,0.2)",
								zIndex: 10,
							}}>
							<Trash2 onClick={handleDeleteEdge} className="text-red-500 cursor-pointer" size={16} />
						</Panel>
					)}
				</ReactFlow>
			</div>

			{/* Dropdowns */}
			<div className="bg-gray-100 h-1/5 p-4 rounded-lg">
				<div className="flex justify-around gap-2">
					<SourceDropdownOptions disabled={!isEditMode} />
					<ProcessorDropdownOptions disabled={!isEditMode} />
					<DestinationDropdownOptions disabled={!isEditMode} />
				</div>
			</div>
		</>
	);
};

export default PipelineGraphEditor;
