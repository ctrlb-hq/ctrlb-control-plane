import { useMemo } from "react";
import Ajv from "ajv";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { Button } from "@/components/ui/button";
import { SheetFooter, SheetClose, SheetContent } from "@/components/ui/sheet";
import { ArrowBigRightDash } from "lucide-react";
import { customEnumRenderer } from "@/components/pipelines/editor/CustomEnumControl";

interface NodeSidePanelProps {
	title: string;
	description?: string;
	formSchema: any;
	uiSchema: any;
	config: any;
	setConfig: (data: any) => void;
	submitLabel?: string;
	submitDisabled?: boolean;
	onSubmit: () => void;
	onDiscard: () => void;
	onDelete?: () => void;
	showDelete?: boolean;
	onErrorsChange?: (errors: any[] | undefined) => void;
}

const applySchemaDefaults = (schema: any, data: any) => {
	const ajv = new Ajv({ useDefaults: true, allErrors: true });
	const validate = ajv.compile(schema);

	// Clone data so we don't mutate parent's object reference
	const clonedData = { ...data };
	validate(clonedData);

	return clonedData;
};

const NodeSidePanel: React.FC<NodeSidePanelProps> = ({
	title,
	description,
	formSchema,
	uiSchema,
	config,
	setConfig,
	submitLabel = "Apply",
	submitDisabled = false,
	onSubmit,
	onDiscard,
	onDelete,
	showDelete = false,
	onErrorsChange,
}) => {
	const theme = createTheme({
		components: {
			MuiFormControl: {
				styleOverrides: {
					root: {
						marginBottom: "0.5rem", // your existing setting
					},
				},
			},
			MuiInputBase: {
				styleOverrides: {
					root: {
						fontSize: "0.8rem", // ~13px
						minHeight: "32px", // compact input height
					},
					input: {
						padding: "6px 8px", // compact padding inside input
					},
				},
			},
			MuiFormLabel: {
				styleOverrides: {
					root: {
						fontSize: "0.75rem", // ~12px label
					},
				},
			},
			MuiSelect: {
				styleOverrides: {
					root: {
						fontSize: "0.8rem", // ~13px select font
					},
				},
			},
			MuiTypography: {
				styleOverrides: {
					h5: {
						fontSize: "1.25rem", // ~16px — now matches your Sheet heading better
						fontWeight: 500, // optional: make it bold like your other headings
					},
				},
			},
		},
	});

	const renderers = [...materialRenderers, customEnumRenderer];

	const configWithDefaults = useMemo(() => {
		return applySchemaDefaults(formSchema, config);
	}, [formSchema, config]);

	return (
		<SheetContent className="w-[36rem] h-full">
			<div className="flex flex-col h-full p-4 gap-4">
				<div className="flex gap-3 items-center">
					<ArrowBigRightDash className="w-6 h-6" /> {/* Slightly bigger arrow */}
					<h2 className="text-2xl font-bold">{title}</h2> {/* Bigger heading */}
				</div>

				{description && (
					<p className="text-gray-500 text-sm">
						{" "}
						{/* Description stays small */}
						{description} <span className="text-blue-500 underline cursor-pointer">Documentation</span>
					</p>
				)}

				<ThemeProvider theme={theme}>
					<div className="flex-grow overflow-y-auto pt-2">
						<div className="p-3 text-xs">
							<JsonForms
								data={configWithDefaults}
								schema={formSchema}
								uischema={uiSchema}
								renderers={renderers}
								cells={materialCells}
								onChange={({ data, errors }) => {
									setConfig(data);
									if (onErrorsChange) {
										onErrorsChange(errors);
									}
								}}
							/>
						</div>
					</div>
				</ThemeProvider>

				<SheetFooter className="pt-4">
					<SheetClose>
						<div className="flex gap-3">
							<Button className="bg-blue-500 text-sm" onClick={onSubmit} disabled={submitDisabled}>
								{submitLabel}
							</Button>
							<Button variant={"outline"} className="text-sm" onClick={onDiscard}>
								Discard Changes
							</Button>
							{showDelete && (
								<Button variant={"outline"} className="text-sm" onClick={onDelete}>
									Delete Node
								</Button>
							)}
						</div>
					</SheetClose>
				</SheetFooter>
			</div>
		</SheetContent>
	);
};

export default NodeSidePanel;
