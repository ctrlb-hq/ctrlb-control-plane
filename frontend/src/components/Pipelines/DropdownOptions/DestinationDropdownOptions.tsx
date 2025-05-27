import { Button } from "@/components/ui/button";
import {
	DropdownMenu,
	DropdownMenuContent,
	DropdownMenuGroup,
	DropdownMenuItem,
	DropdownMenuLabel,
	DropdownMenuSeparator,
	DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";

import { Sheet, SheetClose, SheetContent, SheetFooter } from "@/components/ui/sheet";
import React, { useEffect, useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { TransporterService } from "@/services/transporterService";

import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { customEnumRenderer } from "./CustomEnumControl";


interface destination {
	name: string;
	display_name: string;
	type: string;
	supported_signals: string[];
}
const DestinationDropdownOptions = React.memo(({ disabled }: { disabled: boolean }) => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [destinationOptionValue, setDestinationOptionValue] = useState("");
	const { addChange } = usePipelineChangesLog();
	const [destinations, setDestinations] = useState<destination[]>([]);
	const [data, setData] = useState<object>();
	const [form, setForm] = useState<object>({});
	const [pluginName, setPluginName] = useState();
	const [submitDisabled, setSubmitDisabled] = useState(true);
	const { addNode } = useGraphFlow();
	const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({ type: "VerticalLayout", elements: [] });

	const handleSheetOpen = (e: any) => {
		setPluginName(e);
		setIsSheetOpen(!isSheetOpen);
		handleGetDestinationForm(e);
	};

	const handleSubmit = () => {
		const supported_signals = destinations.find(s => s.name == pluginName)?.supported_signals;

		const newNode = {
			type: "destination",
			position: { x: 0, y: 0 },
			data: {
				type: "exporter",
				name: destinationOptionValue,
				supported_signals: supported_signals,
				component_name: pluginName,
				config: data,
			},
		};

		const newNodeId = addNode(newNode);

		const log = {
			type: "destination",
			id: newNodeId,
			name: destinationOptionValue,
			status: "added",
			initialConfig: undefined,
			finalConfig: data,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));
		setIsSheetOpen(false);
	};

	const handleGetDestination = async () => {
		const res = await TransporterService.getTransporterService("exporter");
		setDestinations(res);
	};

	const handleGetDestinationForm = async (destinationOptionValue: string) => {
		const res = await TransporterService.getTransporterForm(destinationOptionValue);
		const ui=await TransporterService.getTransporterUiSchema(destinationOptionValue);
		setUiSchema(ui);
		setForm(res);
	};

	useEffect(() => {
		handleGetDestination();
	}, [isSheetOpen]);

	const theme = createTheme({
		components: {
			MuiSelect: {
				defaultProps: {
					MenuProps: {
						disablePortal: true,
						container: () => document.body,
					},
				},
			},
		}
	});

	const renderers = [
		...materialRenderers,
		customEnumRenderer
	];

	return (
		<>
			<DropdownMenu>
				<DropdownMenuContent className="w-56">
					<DropdownMenuLabel>Add Destination</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						{destinations.map((destination, index) => (
							<DropdownMenuItem
								key={index}
								onClick={() => {
									handleSheetOpen(destination.name);
									setDestinationOptionValue(destination.display_name);
								}}
							>
								{destination.display_name}
							</DropdownMenuItem>
						))}
					</DropdownMenuGroup>
				</DropdownMenuContent>
				<DropdownMenuTrigger asChild disabled={disabled}>
					<div className="flex justify-center items-center">
						<div className="bg-green-600 h-6 rounded-bl-lg rounded-tl-lg w-2" />
						<div
							className={
								disabled
									? "bg-gray-300 cursor-not-allowed rounded-md shadow-md p-3 border-2 border-gray-300 flex items-center justify-center"
									: "bg-white cursor-pointer rounded-md shadow-md p-3 border-2 border-gray-300 flex items-center justify-center"
							}
							draggable
						>
							Add Destination
						</div>
					</div>
				</DropdownMenuTrigger>
			</DropdownMenu>
			{isSheetOpen && (
				<Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
					<SheetContent className="w-[36rem]">
						<div className="flex flex-col gap-4 p-4">
							<div className="flex gap-3 items-center">
								<p className="text-lg bg-gray-500 items-center rounded-lg p-2 px-3 m-1 text-white">→|</p>
								<h2 className="text-xl font-bold">{destinationOptionValue}</h2>
							</div>
							<p className="text-gray-500">
								Generate the defined log type at the rate desired{" "}
								<span className="text-blue-500 underline">Documentation</span>
							</p>
							<ThemeProvider theme={theme}>
								<div className="mt-3">
									<div className="p-3 ">
										<div className="overflow-y-auto h-[32rem] pt-3">
											{isSheetOpen && form && <JsonForms
												data={data}
												schema={form}
												renderers={renderers}
												uischema={uiSchema}
												cells={materialCells}
												onChange={({ data, errors }) => {
													setData(data);
													const hasErrors = errors && errors.length > 0;
													setSubmitDisabled(!!hasErrors);
												}}
											/>}
										</div>
									</div>
								</div>
							</ThemeProvider>
							<SheetFooter>
								<SheetClose>
									<div className="flex gap-3">
										<Button className="bg-blue-500" onClick={handleSubmit} disabled={submitDisabled}>
											Add Destination
										</Button>
										<Button variant={"outline"} onClick={() => setIsSheetOpen(false)}>
											Discard Changes
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
});

export default DestinationDropdownOptions;
