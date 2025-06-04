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
import { TransporterService } from "@/services/transporter";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { ThemeProvider, createTheme } from "@mui/material/styles";
import { customEnumRenderer } from "@/components/pipelines/editor/CustomEnumControl";
import { JsonSchema } from "@jsonforms/core";
import { ArrowBigRightDash } from "lucide-react";

interface Plugin {
	name: string;
	display_name: string;
	type: string;
	supported_signals: string[];
}

interface Props {
	kind: "receiver" | "processor" | "exporter";
	nodeType: "source" | "processor" | "destination";
	label: string;
	dataType: "receiver" | "exporter";
	disabled: boolean;
}

const PluginDropdownOptions = React.memo(({ kind, nodeType, label, dataType, disabled }: Props) => {
	const [isSheetOpen, setIsSheetOpen] = useState(false);
	const [optionValue, setOptionValue] = useState("");
	const [pluginName, setPluginName] = useState<string | undefined>();
	const [plugins, setPlugins] = useState<Plugin[]>([]);
	const [form, setForm] = useState<JsonSchema>({});
	const [config, setConfig] = useState<object>({});
	const [submitDisabled, setSubmitDisabled] = useState(true);
	const [uiSchema, setUiSchema] = useState<{ type: string; elements: any[] }>({
		type: "VerticalLayout",
		elements: [],
	});

	const { addNode } = useGraphFlow();
	const { addChange } = usePipelineChangesLog();

	const handleSheetOpen = (plugin: string, displayName: string) => {
		setPluginName(plugin);
		setOptionValue(displayName);
		setConfig({});
		setForm({});
		setIsSheetOpen(true);
		fetchForm(plugin);
	};

	const fetchForm = async (plugin: string) => {
		const schema = await TransporterService.getTransporterForm(plugin);
		const ui = await TransporterService.getTransporterUiSchema(plugin);
		setForm(schema);
		setUiSchema(ui);
	};

	const handleSubmit = () => {
		const supported_signals = plugins.find(p => p.name === pluginName)?.supported_signals;
		const newNode = {
			type: nodeType,
			position: { x: 0, y: 0 },
			data: {
				type: dataType,
				name: optionValue,
				supported_signals,
				component_name: pluginName,
				config,
			},
		};
		const id = addNode(newNode);
		const log = {
			type: nodeType,
			id,
			name: optionValue,
			status: "added",
			initialConfig: undefined,
			finalConfig: config,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		localStorage.setItem("changesLog", JSON.stringify([...existingLog, log]));
		setIsSheetOpen(false);
	};

	const fetchPlugins = async () => {
		const res = await TransporterService.getTransporterService(kind);
		setPlugins(res);
	};

	useEffect(() => {
		fetchPlugins();
	}, [isSheetOpen]);

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

	const renderers = [...materialRenderers, customEnumRenderer];

	return (
		<>
			<DropdownMenu>
				<DropdownMenuTrigger asChild disabled={disabled}>
					<Button
						variant="outline"
						className={`flex items-center gap-2 border-2 rounded-md shadow-md px-4 py-2 ${disabled ? "cursor-not-allowed opacity-60" : "hover:bg-muted"
							}`}>
						➕ Add {label}
					</Button>
				</DropdownMenuTrigger>

				<DropdownMenuContent className="w-64 mt-2 shadow-lg border rounded-md bg-white">
					<DropdownMenuLabel className="text-md font-semibold text-gray-700">
						Select a {label}
					</DropdownMenuLabel>
					<DropdownMenuSeparator />
					<DropdownMenuGroup>
						{plugins.map((plugin, index) => (
							<DropdownMenuItem
								key={index}
								onClick={() => handleSheetOpen(plugin.name, plugin.display_name)}
								className="hover:bg-gray-100 cursor-pointer px-3 py-2 text-sm text-gray-800">
								{plugin.display_name}
							</DropdownMenuItem>
						))}
					</DropdownMenuGroup>
				</DropdownMenuContent>
			</DropdownMenu>

			{isSheetOpen && (
				<Sheet open={isSheetOpen} onOpenChange={setIsSheetOpen}>
					<SheetContent className="w-[36rem]">
						<div className="flex flex-col gap-4 p-4">
							<div className="flex gap-3 items-center">
								<ArrowBigRightDash className="w-6 h-6" />
								<h2 className="text-xl font-bold">{optionValue}</h2>
							</div>
							<p className="text-gray-500">
								Generate the defined log type at the rate desired{" "}
								<span className="text-blue-500 underline">Documentation</span>
							</p>
							<ThemeProvider theme={theme}>
								<div className="mt-3">
									<div className="p-3">
										<div className="overflow-y-auto h-[32rem] pt-3">
											{form && (
												<JsonForms
													data={config}
													schema={form}
													uischema={uiSchema}
													renderers={renderers}
													cells={materialCells}
													onChange={({ data, errors }) => {
														setConfig(data);
														setSubmitDisabled(!!errors?.length);
													}}
												/>
											)}
										</div>
									</div>
								</div>
							</ThemeProvider>
							<SheetFooter>
								<SheetClose>
									<div className="flex gap-3">
										<Button className="bg-blue-500" onClick={handleSubmit} disabled={submitDisabled}>
											{`Add ${label}`}
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

export default PluginDropdownOptions;
