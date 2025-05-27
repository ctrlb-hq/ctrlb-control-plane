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
import React, { useEffect, useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { TransporterService } from "@/services/transporterService";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";

interface Processor {
	name: string;
	display_name: string;
	type: string;
	supported_signals: string[];
}

const uiSchema = {
	type: 'VerticalLayout',
	elements: [
		{
			type: 'Control',
			scope: '#/properties/actions',
			label: 'Actions',
			options: {
				detail: {
					type: 'VerticalLayout',
					elements: [
						{
							type: 'Control',
							scope: '#/properties/key'
						},
						{
							type: 'Control',
							scope: '#/properties/action',

						},
						{
							type: 'Control',
							scope: '#/properties/from_attribute'
						},
						{
							type: 'Control',
							scope: '#/properties/value'
						}
					]
				}
			}
		},
		{
			type: 'Group',
			label: 'Include Filter',
			elements: [
				{
					type: 'Control',
					scope: '#/properties/include/properties/match_type',

				},
				{
					type: 'Control',
					scope: '#/properties/include/properties/attributes',
					options: {
						detail: {
							type: 'VerticalLayout',
							elements: [
								{ type: 'Control', scope: '#/properties/key' },
								{ type: 'Control', scope: '#/properties/value' }
							]
						}
					}
				}
			]
		},
		{
			type: 'Group',
			label: 'Exclude Filter',
			elements: [
				{
					type: 'Control',
					scope: '#/properties/exclude/properties/match_type',

				},
				{
					type: 'Control',
					scope: '#/properties/exclude/properties/attributes',

				}
			]
		}
	]
};


import { ThemeProvider, createTheme } from "@mui/material/styles";
import { customEnumRenderer } from "./CustomEnumControl";
const ProcessorDropdownOptions = React.memo(({ disabled }: { disabled: boolean }) => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [processorOptionValue, setProcessorOptionValue] = useState("");
	const { addChange } = usePipelineChangesLog();
	const [form, setForm] = useState<object>({});
	const [config, setConfig] = useState<object>({});
	const [pluginName, setPluginName] = useState();
	const [processors, setProcessors] = useState<Processor[]>([]);
	const [submitDisabled, setSubmitDisabled] = useState(true);
	const { addNode } = useGraphFlow();

	const handleSheetOpen = (e: any) => {
		setPluginName(e);
		setConfig({});
		setForm({});
		setIsSheetOpen(true);
		handleGetProcessorForm(e);
	};

	const handleSubmit = () => {
		const supported_signals = processors.find(s => s.name == pluginName)?.supported_signals;

		const newNode = {
			type: "processor",
			position: { x: 0, y: 0 },
			data: {
				type: "receiver",
				name: processorOptionValue,
				supported_signals: supported_signals,
				component_name: pluginName,
				config: config,
			},
		};

		const newNodeId = addNode(newNode);

		const log = {
			type: "processor",
			id: newNodeId,
			name: processorOptionValue,
			status: "added",
			initialConfig: undefined,
			finalConfig: config,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		const updatedLog = [...existingLog, log];
		localStorage.setItem("changesLog", JSON.stringify(updatedLog));

		setIsSheetOpen(false);
	};

	const handleGetProcessor = async () => {
		const res = await TransporterService.getTransporterService("processor");
		setProcessors(res);
	};

	const handleGetProcessorForm = async (processorOptionValue: string) => {
		const res = await TransporterService.getTransporterForm(processorOptionValue);
		console.log(res);
		setForm(res);
	};

	useEffect(() => {
		handleGetProcessor();
	}, []);

	const theme = createTheme({
		components: {
			MuiSelect: {
				defaultProps: {
					MenuProps: {
						container: document.body,
						disablePortal: true,
					},
				},
			},

		},
	});

	const renderers:any = [
		...materialRenderers,
		customEnumRenderer
	];

	return (
		<>
			<DropdownMenu>
				<DropdownMenuContent className="w-56">
					<DropdownMenuLabel>Add Processor</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						<DropdownMenuSub>
							{processors!.map((processor, index) => (
								<DropdownMenuItem
									key={index}
									onClick={() => {
										handleSheetOpen(processor.name);
										setProcessorOptionValue(processor.display_name);
									}}
								>
									{processor.display_name}
								</DropdownMenuItem>
							))}
						</DropdownMenuSub>
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
							Add Processor
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
								<h2 className="text-xl font-bold">{processorOptionValue}</h2>
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
												data={config}
												schema={form}
												uischema={uiSchema}
												renderers={renderers}
												cells={materialCells}
												onChange={({ data, errors }) => {
													setConfig(data);
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
											Add Processor
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

export default ProcessorDropdownOptions;
