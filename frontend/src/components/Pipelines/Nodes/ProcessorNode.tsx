import { Handle, Position } from "reactflow";
import { Sheet, SheetTrigger, SheetClose, SheetContent, SheetFooter } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { useNodeValue } from "@/context/useNodeContext";
import { useEffect, useState } from "react";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { JsonForms } from "@jsonforms/react";

import { materialCells, materialRenderers } from "@jsonforms/material-renderers";

import { ThemeProvider, createTheme } from "@mui/material/styles";
import { TransporterService } from "@/services/transporterService";
import { ArrowBigRightDash } from "lucide-react";

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

const renderers = [...materialRenderers];

export const ProcessorNode = ({ data: Data }: any) => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const { setNodeValue } = useNodeValue();
	const { addChange } = usePipelineChangesLog();
	const getSource = JSON.parse(localStorage.getItem("Nodes") || "[]").find(
		(source: any) => source.component_name === Data.component_name,
	);
	const processorConfig = getSource?.config;
	const [data, setData] = useState<object>(Data.config);
	const [form, setForm] = useState<FormSchema>({});

	const ProcessorLabel = Data.supported_signals;
	const handleSubmit = () => {
		const log = { type: "processor", name: Data.name, status: "edited" };
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		const nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
		const updatedNodes = nodes.map((node: any) =>
			node.component_name === Data.component_name ? { ...node, config: data } : node,
		);
		localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
		setIsSheetOpen(false);
	};

	const getForm = async () => {
		const res = await TransporterService.getTransporterForm(Data.component_name);
		setForm(res);
	};

	useEffect(() => {
		getForm();
	}, []);

	const handleDeleteNode = () => {
		// Delete node
		setNodeValue(prev => prev.filter(node => node.id !== Data.component_id));
		setNodeValue(prev => prev.filter(node => node.id !== Data.id));

		// Delete connected edges
		const edges = JSON.parse(localStorage.getItem("PipelineEdges") || "[]");
		const updatedEdges = edges.filter(
			(edge: any) => edge.source !== Data.id && edge.target !== Data.id,
		);
		localStorage.setItem("PipelineEdges", JSON.stringify(updatedEdges));

		// Log the change
		const log = { type: "processor", name: Data.name, status: "deleted" };
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		// Update nodes in storage
		const nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
		const updatedNodes = nodes.filter((node: any) => node.component_name !== Data.component_name);
		localStorage.setItem("Nodes", JSON.stringify(updatedNodes));

		setIsSheetOpen(false);
	};

	return (
		<Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
			<SheetTrigger asChild>
				<div onClick={() => setIsSheetOpen(true)} className="flex items-center">
					<div className="bg-green-600 h-6 rounded-tl-lg rounded-bl-lg w-2" />
					<div className="bg-gray-200 rounded  p-4 h-[4rem]  w-[7.5rem] px-2 py-1 flex flex-col items-center justify-center relative text-center">
						<Handle
							type="target"
							position={Position.Left}
							className="bg-green-600 w-0 h-0 rounded-full"
						/>

						<div className="text-[9px] leading-3 font-medium break-words max-w-full">
							{Data.name}
						</div>
						<div className="flex flex-wrap justify-center gap-1 text-[8px] mt-1 text-gray-700">
							{ProcessorLabel &&
								ProcessorLabel.map((source: any, index: number) => (
									<p style={{ fontSize: "8px" }} key={index}>
										{source}
									</p>
								))}
						</div>
						<Handle
							type="source"
							position={Position.Right}
							className="bg-green-600 w-0 h-0 rounded-full"
						/>
					</div>
					<div className="bg-green-600 h-6 rounded-tr-lg rounded-br-lg w-2" />
				</div>
			</SheetTrigger>
			<SheetContent className="w-[36rem]">
				<div className="flex flex-col gap-4 p-4">
					<div className="flex gap-3 items-center">
						{/* <p className="text-lg bg-gray-500 items-center rounded-lg p-2 px-3 m-1 text-white">→|</p> */}
						<ArrowBigRightDash className="w-6 h-6" />
						<h2 className="text-xl font-bold">{Data.name}</h2>
					</div>
					<p className="text-gray-500">
						Generate the defined log type at the rate desired{" "}
						<span className="text-blue-500 underline">Documentation</span>
					</p>
					<ThemeProvider theme={theme}>
						<div className="mt-3">
							<div className="text-2xl p-4 font-semibold bg-gray-100">{form.title}</div>
							<div className="p-3 ">
								<div className="overflow-y-auto h-[29rem]">
									<JsonForms
										data={data}
										schema={form}
										renderers={renderers}
										cells={materialCells}
										onChange={({ data }) => setData(data)}
									/>
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
								<Button variant={"outline"} onClick={() => setIsSheetOpen(false)}>
									Discard Changes
								</Button>
								<Button variant={"outline"} onClick={handleDeleteNode}>
									Delete Node
								</Button>
							</div>
						</SheetClose>
					</SheetFooter>
				</div>
			</SheetContent>
		</Sheet>
	);
};
