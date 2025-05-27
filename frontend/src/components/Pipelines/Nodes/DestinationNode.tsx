import { Handle, Position } from "reactflow";
import { Sheet, SheetClose, SheetContent, SheetFooter, SheetTrigger } from "@/components/ui/sheet";
import React, { useEffect, useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { Button } from "../../ui/button";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { TransporterService } from "@/services/transporterService";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { ArrowBigRightDash } from "lucide-react";
import { customEnumRenderer } from "../DropdownOptions/CustomEnumControl";

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

const renderers = [
	...materialRenderers,
	customEnumRenderer
];

export const DestinationNode = React.memo(({ data: Data }: any) => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const { deleteNode, updateNodeConfig } = useGraphFlow();
	const { addChange } = usePipelineChangesLog();
	const [form, setForm] = useState<FormSchema>({});

	const DestinationLabel = Data.supported_signals;
	const handleSubmit = () => {
		const log = {
			type: "destination",
			id: Data.component_id,
			name: Data.name,
			status: "edited",
			initialConfig: Data.config,
			finalConfig: config,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		updateNodeConfig(Data.component_id, config);

		const sources = JSON.parse(localStorage.getItem("Destination") || "[]");
		const updatedSources = sources.filter(
			(source: any) => source.component_name !== Data.component_name,
		);
		localStorage.setItem("Destination", JSON.stringify(updatedSources));

		setIsSheetOpen(false);
	};

	const getForm = async () => {
		const res = await TransporterService.getTransporterForm(Data.component_name);
		setForm(res as FormSchema);
	};

	useEffect(() => {
		getForm();
	}, []);

	const handleDeleteNode = () => {
		deleteNode(Data.component_id);
		const log = {
			type: "destination",
			id: Data.component_id,
			name: Data.name,
			status: "deleted",
			initialConfig: Data.config,
			finalConfig: undefined,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));
		setIsSheetOpen(false);
	};

	// const getSource = JSON.parse(localStorage.getItem("Nodes") || "[]").find(
	// 	(source: any) => source.component_name === Data.component_name,
	// );
	// const sourceConfig = getSource?.config;

	const [config, setConfig] = useState<object>(Data.config);

	return (
		<Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
			<SheetTrigger asChild>
				<div className="flex items-center">
					<div className="bg-green-600 h-6 rounded-tl-lg rounded-bl-lg w-2" />
					<div className="bg-gray-200 flex justify-between items-center rounded-md h-[4rem] w-[8rem] ">
						<Handle
							type="target"
							position={Position.Left}
							className="bg-green-600 w-0 h-0 rounded-full"
						/>
						<div className="flex ml-5 flex-col w-full">
							<div
								style={{ fontSize: "9px", lineHeight: "0.8rem" }}
								className="font-medium flex justify-start"
							>
								{Data.name}
							</div>
							<div className="flex flex-wrap gap-1 text-[8px] mt-1 text-gray-700">
								{DestinationLabel.map((source: any, index: number) => (
									<p style={{ fontSize: "8px" }} key={index}>
										{source}
									</p>
								))}
							</div>
						</div>
						{Data.label === "ctrlB" ? (
							<div className="flex items-center rounded-br-md rounded-tr-md bg-green-500 h-[4rem]">
								<div className="bg-white rounded-md m-1">
									<img src="./ctrlb-logo.png" width={"48px"} />
								</div>
							</div>
						) : (

							<div className="flex items-center justify-center rounded-br-md rounded-tr-md bg-gray-500 h-[4rem] w-[3rem]">
								<ArrowBigRightDash className="text-white w-6 h-6" />
							</div>
						)}
					</div>
				</div>
			</SheetTrigger>
			<SheetContent className="w-[36rem]">
				<div className="flex flex-col gap-4 p-4">
					<div className="flex flex-col gap-3">
						<div className="flex items-center gap-4">
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
									<div className="overflow-y-auto h-[32rem] pt-3">
										{form && isSheetOpen && <JsonForms
											data={config}
											schema={form}
											renderers={renderers}
											cells={materialCells}
											onChange={({ data }) => setConfig(data)}
										/>}
									</div>
								</div>
							</div>
						</ThemeProvider>
					</div>
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
});
