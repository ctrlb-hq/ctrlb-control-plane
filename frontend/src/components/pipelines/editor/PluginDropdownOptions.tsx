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
import { Sheet } from "@/components/ui/sheet";
import React, { useEffect, useState } from "react";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { ComponentService } from "@/services/component";
import { JsonSchema } from "@jsonforms/core";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";

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
		const schema = await ComponentService.getTransporterForm(plugin);
		const ui = await ComponentService.getTransporterUiSchema(plugin);
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
			component_type: pluginName,
		};
		const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
		addChange(log);
		localStorage.setItem("changesLog", JSON.stringify([...existingLog, log]));
		setIsSheetOpen(false);
	};

	const fetchPlugins = async () => {
		const res = await ComponentService.getTransporterService(kind);
		setPlugins(res);
	};

	useEffect(() => {
		fetchPlugins();
	}, [isSheetOpen]);

	return (
		<>
			<DropdownMenu>
				<DropdownMenuTrigger asChild disabled={disabled}>
					<Button
						variant="outline"
						className={`flex items-center gap-2 border-2 rounded-md shadow-md px-4 py-2 ${
							disabled ? "cursor-not-allowed opacity-60" : "hover:bg-muted"
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
					<NodeSidePanel
						title={optionValue}
						formSchema={form}
						uiSchema={uiSchema}
						config={config}
						setConfig={setConfig}
						submitLabel={`Add ${label}`}
						submitDisabled={submitDisabled}
						onSubmit={handleSubmit}
						onDiscard={() => setIsSheetOpen(false)}
						showDelete={false}
					/>
				</Sheet>
			)}
		</>
	);
});

export default PluginDropdownOptions;
