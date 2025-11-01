import React, { useEffect, useState } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import { Sheet, SheetTrigger } from "@/components/ui/sheet";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { ComponentService } from "@/services/component";
import { ArrowBigRightDash } from "lucide-react";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";

interface FormSchema {
	title?: string;
	type?: string;
	properties?: Record<string, any>;
	required?: string[];
	[key: string]: any;
}

interface GenericNodeProps extends NodeProps {
	type: "source" | "processor" | "destination";
	triggerPosition?: "left" | "right";
	labelComponent?: React.ReactNode;
}

const GenericNode = React.memo(({ data: Data, type }: GenericNodeProps) => {
	const [isOpen, setIsOpen] = useState(false);
	const { deleteNode, updateNodeConfig } = useGraphFlow();
	const [form, setForm] = useState<FormSchema>({});
	const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({
		type: "VerticalLayout",
		elements: [],
	});
	const [config, setConfig] = useState<object>(Data.config);
	const nodeId = Data.component_id.toString();
	
	useEffect(() => {
		setConfig(Data.config || {});
	}, [Data.config]);

	const handleDeleteNode = () => {
		deleteNode(nodeId);
		setIsOpen(false);
	};

	const handleSubmit = (submittedConfig: any) => {
		updateNodeConfig(nodeId, submittedConfig);
		setIsOpen(false);
	};

	const getForm = async () => {
		const res = await ComponentService.getTransporterForm(Data.component_name);
		const ui = await ComponentService.getTransporterUiSchema(Data.component_name);
		setForm(res);
		setUiSchema(ui);
	};

	useEffect(() => {
		if (isOpen) getForm();
	}, [isOpen]);

	return (
		<Sheet open={isOpen} onOpenChange={setIsOpen}>
			<SheetTrigger asChild>
				<div onClick={() => setIsOpen(true)} className="flex items-center gap-0">
					{type === "source" && (
						<>
							<div className="bg-gray-500 h-[4rem] w-[2rem] rounded-l-md flex items-center justify-center">
								<ArrowBigRightDash className="text-white w-6 h-6" />
							</div>
							<div className="bg-gray-200 flex flex-col items-center justify-center h-[4rem] w-[7rem] rounded-r-md relative">
								<Handle
									type="source"
									position={Position.Right}
									className="bg-green-600 w-1.5 h-3.5 rounded-full"
									style={{ right: "-6px", top: "50%", transform: "translateY(-50%)" }}
								/>
								<div className="text-[9px] font-medium text-center break-words max-w-full">{Data.name}</div>
								<div className="flex flex-wrap justify-center gap-1 text-[8px] mt-1 text-gray-700">
									{Data.supported_signals?.map((sig: string, idx: number) => <p key={idx}>{sig}</p>)}
								</div>
							</div>
						</>
					)}

					{type === "processor" && (
						<div className="bg-gray-200 flex flex-col items-center justify-center h-[4rem] w-[7rem] rounded-md relative">
							<Handle
								type="target"
								position={Position.Left}
								className="!bg-green-600 w-2 h-4 rounded-full border-2 border-white"
								style={{ left: "-6px", top: "50%", transform: "translateY(-50%)" }}
							/>
							<Handle
								type="source"
								position={Position.Right}
								className="!bg-green-600 w-2 h-4 rounded-full border-2 border-white"
								style={{ right: "-6px", top: "50%", transform: "translateY(-50%)" }}
							/>
							<div className="text-[9px] font-medium text-center break-words max-w-full">{Data.name}</div>
							<div className="flex flex-wrap justify-center gap-1 text-[8px] mt-1 text-gray-700">
								{Data.supported_signals?.map((sig: string, idx: number) => <p key={idx}>{sig}</p>)}
							</div>
						</div>
					)}

					{type === "destination" && (
						<>
							<div className="bg-gray-200 flex flex-col items-center justify-center h-[4rem] w-[7rem] rounded-l-md relative">
								<Handle
									type="target"
									position={Position.Left}
									className="!bg-green-600 w-2 h-4 rounded-full border-2 border-white"
									style={{ left: "-6px", top: "50%", transform: "translateY(-50%)" }}
								/>
								<div className="text-[9px] font-medium text-center break-words max-w-full">{Data.name}</div>
								<div className="flex flex-wrap justify-center gap-1 text-[8px] mt-1 text-gray-700">
									{Data.supported_signals?.map((sig: string, idx: number) => <p key={idx}>{sig}</p>)}
								</div>
							</div>
							<div className="bg-gray-500 h-[4rem] w-[2rem] rounded-r-md flex items-center justify-center">
								<ArrowBigRightDash className="text-white w-6 h-6" />
							</div>
						</>
					)}
				</div>
			</SheetTrigger>

			<NodeSidePanel
				title={Data.name}
				formSchema={form}
				uiSchema={uiSchema}
				config={config}
				setConfig={setConfig}
				submitLabel="Apply"
				onSubmit={handleSubmit}
				onDiscard={() => setIsOpen(false)}
				onDelete={handleDeleteNode}
				showDelete={true}
				isOpen={isOpen}
			/>
		</Sheet>
	);
});

export default GenericNode;
