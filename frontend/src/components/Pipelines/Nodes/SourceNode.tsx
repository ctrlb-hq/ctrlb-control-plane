import { useEffect, useState } from "react";
import { Handle, Position } from "reactflow";
import { Sheet, SheetClose, SheetContent, SheetFooter, SheetTrigger } from "@/components/ui/sheet";
import { Button } from "@/components/ui/button";
import { useGraphFlow } from "@/context/useGraphFlowContext";
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
export const SourceNode = ({ data: Data }: any) => {
	const [isSidebarOpen, setIsSidebarOpen] = useState(false);
	const { updateNodes, deleteNode } = useGraphFlow();
	const { addChange } = usePipelineChangesLog();
	const [form, setForm] = useState<FormSchema>({});
	const SourceLabel = Data.supported_signals || "";

	const handleDeleteNode = () => {
		deleteNode(Data.id);
		setIsSidebarOpen(false);
	};

	const getForm = async () => {
		const res = await TransporterService.getTransporterForm(Data.component_name);
		setForm(res);
	};

	const getSource = JSON.parse(localStorage.getItem("Nodes") || "[]").find(
		(source: any) => source.component_name === Data.component_name,
	);
	const sourceConfig = getSource?.config;
	const [data, setData] = useState<object>(Data.config);

	useEffect(() => {
		getForm();
	}, []);

	const handleSubmit = () => {
		const log = { type: "source", name: Data.name, status: "edited" };
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		const nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
		const updatedNodes = nodes.map((node: any) =>
			node.component_name === Data.component_name ? { ...node, config: data } : node,
		);
		localStorage.setItem("Nodes", JSON.stringify(updatedNodes));

		setIsSidebarOpen(false);
	};
	return (
		<Sheet open={isSidebarOpen} onOpenChange={setIsSidebarOpen}>
			<SheetTrigger asChild>
				<div onClick={() => setIsSidebarOpen(true)} className="flex items-center">
					<div className="flex items-center justify-center rounded-bl-md rounded-tl-md bg-gray-500 h-[4rem] w-[2rem]">
						<ArrowBigRightDash className="text-white w-6 h-6" />
					</div>
					<div className="bg-gray-200 rounded-tr-md rounded-br-md  px-2 py-2 h-[4rem]  ">
						<div className="text-[9px] leading-3 font-medium text-center break-words max-w-[5.5rem]">
							{Data.name}
						</div>
						<div className="flex flex-wrap justify-center gap-1 text-[8px] mt-1 text-gray-700">
							{SourceLabel &&
								SourceLabel.map((source: any, index: number) => (
									<p className="text-[8px] mt-1" key={index}>
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
						<SheetFooter>
							<SheetClose>
								<div className="flex gap-3">
									<Button className="bg-blue-500" onClick={handleSubmit}>
										Apply
									</Button>
									<Button variant={"outline"} onClick={() => setIsSidebarOpen(false)}>
										Discard Changes
									</Button>
									<Button variant={"outline"} onClick={handleDeleteNode}>
										Delete Node
									</Button>
								</div>
							</SheetClose>
						</SheetFooter>
					</ThemeProvider>
				</div>
			</SheetContent>
		</Sheet>
	);
};
