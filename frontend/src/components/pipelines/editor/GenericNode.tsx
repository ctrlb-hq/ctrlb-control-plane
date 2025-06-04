import React, { useEffect, useState } from "react";
import { Handle, Position, NodeProps } from "reactflow";
import { Sheet, SheetTrigger, SheetClose, SheetContent, SheetFooter } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { TransporterService } from "@/services/transporterService";
import { ArrowBigRightDash } from "lucide-react";
import { customEnumRenderer } from "./CustomEnumControl";

interface FormSchema {
	title?: string;
	type?: string;
	properties?: Record<string, any>;
	required?: string[];
	[key: string]: any;
}

const theme = createTheme({
	components: {
		MuiFormControl: {
			styleOverrides: {
				root: {
					marginBottom: "0.5rem",
				},
			},
		},
	},
});

const renderers = [...materialRenderers, customEnumRenderer];

interface GenericNodeProps extends NodeProps {
	type: "source" | "processor" | "destination";
	triggerPosition?: "left" | "right";
	labelComponent?: React.ReactNode;
}

const GenericNode = React.memo(({ data: Data, type, labelComponent }: GenericNodeProps) => {
	const [isOpen, setIsOpen] = useState(false);
	const { deleteNode, updateNodeConfig } = useGraphFlow();
	const { addChange } = usePipelineChangesLog();
	const [form, setForm] = useState<FormSchema>({});
	const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({
		type: "VerticalLayout",
		elements: [],
	});
	const [config, setConfig] = useState<object>(Data.config);
	const nodeId = Data.component_id.toString();

	const handleDeleteNode = () => {
		const log = {
			type,
			id: nodeId,
			name: Data.name,
			status: "deleted",
			initialConfig: Data.config,
			finalConfig: undefined,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));
		deleteNode(nodeId);
		setIsOpen(false);
	};

	const handleSubmit = () => {
		const log = {
			type,
			id: nodeId,
			name: Data.name,
			status: "edited",
			initialConfig: Data.config,
			finalConfig: config,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));
		updateNodeConfig(nodeId, config);
		setIsOpen(false);
	};

	const getForm = async () => {
		const res = await TransporterService.getTransporterForm(Data.component_name);
		const ui = await TransporterService.getTransporterUiSchema(Data.component_name);
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
					{/* === Source Node === */}
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

					{/* === Processor Node === */}
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

					{/* === Destination Node === */}
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

			<SheetContent className="w-[36rem]">
				<div className="flex flex-col gap-4 p-4">
					<div className="flex gap-3 items-center">
						<ArrowBigRightDash className="w-6 h-6" />
						<h2 className="text-xl font-bold">{Data.name}</h2>
					</div>
					<p className="text-gray-500">
						Generate the defined log type at the rate desired.{" "}
						<span className="text-blue-500 underline">Documentation</span>
					</p>
					<ThemeProvider theme={theme}>
						<div className="mt-3">
							<div className="text-2xl p-4 font-semibold bg-gray-100">{form.title}</div>
							<div className="p-3">
								<div className="overflow-y-auto h-[32rem] pt-3">
									{form && isOpen && (
										<JsonForms
											data={config}
											schema={form}
											uischema={uiSchema}
											renderers={renderers}
											cells={materialCells}
											onChange={({ data }) => setConfig(data)}
										/>
									)}
								</div>
							</div>
						</div>
					</ThemeProvider>
					<SheetFooter>
						<SheetClose>
							<div className="flex gap-3">
								<Button className="bg-blue-500" onClick={handleSubmit}>
									Apply
								</Button>
								<Button variant="outline" onClick={() => setIsOpen(false)}>
									Discard Changes
								</Button>
								<Button variant="outline" onClick={handleDeleteNode}>
									Delete Node
								</Button>
							</div>
						</SheetClose>
					</SheetFooter>
				</div>
			</SheetContent>
		</Sheet>
	);
});

export default GenericNode;
