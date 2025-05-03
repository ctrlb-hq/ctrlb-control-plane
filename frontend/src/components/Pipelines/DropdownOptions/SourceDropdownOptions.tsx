import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuSub,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Sheet, SheetClose, SheetContent, SheetFooter } from "@/components/ui/sheet";
import { useEffect, useState } from "react";
import { useNodeValue } from "@/context/useNodeContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { TransporterService } from "@/services/transporterService";
import { JsonForms } from "@jsonforms/react";

import { materialCells, materialRenderers } from "@jsonforms/material-renderers";

import { ThemeProvider, createTheme } from "@mui/material/styles";

interface sources {
	name: string;
	display_name: string;
	type: string;
	supported_signals: string[];
}

const SourceDropdownOptions = () => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [sourceOptionValue, setSourceOptionValue] = useState("");
	const { setNodeValue } = useNodeValue();
	const { addChange } = usePipelineChangesLog();
	const [form, setForm] = useState<object>({});
	const [sources, setSources] = useState<sources[]>([]);
	const [data, setData] = useState<object>();
	const [pluginName, setPluginName] = useState();
	const [submitDisabled, setSubmitDisabled] = useState(true);

	const handleSheetOPen = (e: any) => {
		setPluginName(e);
		setIsSheetOpen(!isSheetOpen);
		handleGetSourceForm(e);
	};
	const existingNodes = JSON.parse(localStorage.getItem("Nodes") || "[]");

	const handleSubmit = () => {
		const supported_signals = sources.find(s => s.name == pluginName)?.supported_signals;
		const newNode = {
			id: (existingNodes.length + 1).toString(),
			type: "source",
			position: { x: 350, y: 450 },
			data: {
				label: (
					<div style={{ fontSize: "10px", textAlign: "center" }}>
						{`${sourceOptionValue}-(${existingNodes.length + 1})`}
					</div>
				),
				type: "receiver",
				component_id: (existingNodes.length + 1).toString(),
				name: sourceOptionValue,
				supported_signals: supported_signals,
				component_name: pluginName,
				config: data,
			},
		};

		const nodeToBeAdded = {
			component_id: existingNodes.length + 1,
			component_role: "receiver",
			config: data,
			name: sourceOptionValue,
			component_name: pluginName,
			supported_signals: supported_signals,
		};

		const log = { type: "source", name: sourceOptionValue, status: "added" };
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		localStorage.setItem("Nodes", JSON.stringify([...existingNodes, nodeToBeAdded]));
		setNodeValue(prev => [...prev, newNode]);

		setIsSheetOpen(false);
	};

	const handleGetSources = async () => {
		const res = await TransporterService.getTransporterService("receiver");
		setSources(res);
	};

	const handleGetSourceForm = async (sourceOptionValue: string) => {
		const res = await TransporterService.getTransporterForm(sourceOptionValue);
		setForm(res);
	};

	useEffect(() => {
		handleGetSources();
	}, []);

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

	return (
		<>
			<DropdownMenu>
				<DropdownMenuContent className="w-56">
					<DropdownMenuLabel>Add Source</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						<DropdownMenuSub>
							{sources!.map((source, index) => (
								<DropdownMenuItem
									key={index}
									onClick={() => {
										handleSheetOPen(source.name);
										setSourceOptionValue(source.display_name);
									}}
								>
									{source.display_name}
								</DropdownMenuItem>
							))}
						</DropdownMenuSub>
					</DropdownMenuGroup>
				</DropdownMenuContent>
				<DropdownMenuTrigger asChild>
					<div className="flex justify-center items-center">
						<div
							className="bg-white cursor-pointer rounded-md shadow-md p-3 border-2 border-gray-300 flex items-center justify-center"
							draggable
						>
							Add Source
						</div>
						<div className="bg-green-600 h-6 rounded-tr-lg rounded-br-lg w-2" />
					</div>
				</DropdownMenuTrigger>
			</DropdownMenu>
			{isSheetOpen && (
				<Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
					<SheetContent className="w-[36rem]">
						<div className="flex flex-col gap-4 p-4">
							<div className="flex gap-3 items-center">
								<p className="text-lg bg-gray-500 items-center rounded-lg p-2 px-3 m-1 text-white">→|</p>
								<h2 className="text-xl font-bold">{sourceOptionValue}</h2>
							</div>
							<p className="text-gray-500">
								Generate the defined log type at the rate desired{" "}
								<span className="text-blue-500 underline">Documentation</span>
							</p>
							<ThemeProvider theme={theme}>
								<div className="mt-3">
									<div className="p-3 ">
										<div className="overflow-y-auto h-[32rem]">
											<JsonForms
												data={data}
												schema={form}
												renderers={renderers}
												cells={materialCells}
												onChange={({ data, errors }) => {
													setData(data);
													const hasErrors = errors && errors.length > 0;
													setSubmitDisabled(!!hasErrors);
												}}
											/>
										</div>
									</div>
								</div>
							</ThemeProvider>
							<SheetFooter>
								<SheetClose>
									<div className="flex gap-3">
										<Button className="bg-blue-500" onClick={handleSubmit} disabled={submitDisabled}>
											Apply Changes
										</Button>
										<Button variant={"outline"} onClick={() => setIsSheetOpen(false)}>
											Discard Changes
										</Button>
										<Button variant={"outline"} onClick={() => setIsSheetOpen(false)}>
											Delete Node
										</Button>
									</div>
								</SheetClose>
							</SheetFooter>
						</div>
					</SheetContent>
				</Sheet>
			)}
		</>
	);
};

export default SourceDropdownOptions;
