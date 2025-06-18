import { useEffect, useMemo, useState, useRef } from "react";
import Ajv from "ajv";
import { createTheme, ThemeProvider } from "@mui/material/styles";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { Button } from "@/components/ui/button";
import { SheetFooter, SheetClose, SheetContent } from "@/components/ui/sheet";
import { ArrowBigRightDash } from "lucide-react";
import { customEnumRenderer } from "@/components/pipelines/editor/custom_renderers/CustomEnumControl";
import { customKeyValueRenderer } from "@/components/pipelines/editor/custom_renderers/CustomKeyValueControl";

interface NodeSidePanelProps {
	title: string;
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
	isOpen?: boolean; // for mounted Sheet usage
}

const applySchemaDefaults = (schema: any, data: any) => {
	const ajv = new Ajv({ useDefaults: true, allErrors: true });
	const validate = ajv.compile(schema);

	const clonedData = { ...data };
	validate(clonedData);

	return clonedData;
};

const NodeSidePanel: React.FC<NodeSidePanelProps> = ({
	title,
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
	isOpen = true,
}) => {
	const [showErrors, setShowErrors] = useState(false);
	const [draftConfig, setDraftConfig] = useState(() => applySchemaDefaults(formSchema, config));
	const [formErrors, setFormErrors] = useState<any[]>([]);
	const closeRef = useRef<HTMLButtonElement | null>(null);

	// Reset draft config on sheet open if component is mounted always
	useEffect(() => {
		if (isOpen !== undefined && isOpen) {
			setDraftConfig(applySchemaDefaults(formSchema, config));
			setShowErrors(false);
		}
	}, [isOpen, formSchema, config]);

	const ajv = useMemo(() => new Ajv({ allErrors: true, strict: false }), []);
	const validate = useMemo(() => ajv.compile(formSchema), [ajv, formSchema]);

	const theme = useMemo(
		() =>
			createTheme({
				components: {
					MuiFormControl: { styleOverrides: { root: { marginBottom: "0.5rem" } } },
					MuiInputBase: {
						styleOverrides: {
							root: { fontSize: "0.8rem", minHeight: "32px" },
							input: { padding: "6px 8px" },
						},
					},
					MuiFormLabel: { styleOverrides: { root: { fontSize: "0.75rem" } } },
					MuiSelect: { styleOverrides: { root: { fontSize: "0.8rem" } } },
					MuiTypography: { styleOverrides: { h5: { fontSize: "1rem", fontWeight: 500 } } },
					MuiAccordion: { styleOverrides: { root: { marginBottom: "1rem" } } },
					MuiAvatar: {
						styleOverrides: {
							root: {
								minWidth: "1.8rem",
								width: "1.8rem",
								height: "1.8rem",
								fontSize: "0.8rem",
								marginRight: "0.5rem",
								boxSizing: "border-box",
								backgroundColor: "#3B82F6",
								color: "#FFFFFF",
							},
						},
					},
					MuiGrid: {
						styleOverrides: {
							root: {
								"&.MuiGrid-container.MuiGrid-direction-xs-column": {
									rowGap: "1rem",
								},
							},
						},
					},
				},
			}),
		[],
	);

	const renderers = useMemo(
		() => [...materialRenderers, customEnumRenderer, customKeyValueRenderer],
		[],
	);

	const handleSubmit = () => {
		setShowErrors(true);

		const isValid = validate(draftConfig);
		if (!isValid) {
			setFormErrors(validate.errors || []);
			return;
		}

		setFormErrors([]);
		setConfig(draftConfig);
		onSubmit();
		closeRef.current?.click();
	};

	const handleDiscard = () => {
		setShowErrors(false);
		setDraftConfig(applySchemaDefaults(formSchema, config));
		onDiscard();
	};

	return (
		<SheetContent className="w-[36rem] h-full">
			<div className="flex flex-col h-full p-4 gap-4">
				<div className="flex gap-3 items-center">
					<ArrowBigRightDash className="w-6 h-6" />
					<h2 className="text-2xl font-bold">{title}</h2>
				</div>

				<p className="text-gray-500 text-sm">
					{"For more information please refer "}
					<span className="text-blue-500 underline cursor-pointer">Documentation</span>
				</p>

				<ThemeProvider theme={theme}>
					<div className="flex-grow overflow-y-auto pt-2 p-3 text-xs">
						<JsonForms
							data={draftConfig}
							schema={formSchema}
							uischema={uiSchema}
							renderers={renderers}
							cells={materialCells}
							validationMode={showErrors ? "ValidateAndShow" : "ValidateAndHide"}
							onChange={({ data }) => {
								setDraftConfig(data);
								const valid = validate(data);
								setFormErrors(valid ? [] : validate.errors || []);
							}}
						/>
					</div>
				</ThemeProvider>

				<SheetFooter className="pt-4">
					<SheetClose asChild>
						<button ref={closeRef} className="hidden" />
					</SheetClose>

					<div className="flex gap-3">
						<Button
							className="bg-blue-500 text-sm"
							onClick={handleSubmit}
							disabled={submitDisabled && formErrors.length > 0}>
							{submitLabel}
						</Button>
						<SheetClose asChild>
							<Button variant="outline" className="text-sm" onClick={handleDiscard}>
								Discard Changes
							</Button>
						</SheetClose>
						{showDelete && (
							<SheetClose asChild>
								<Button variant="outline" className="text-sm" onClick={onDelete}>
									Delete Node
								</Button>
							</SheetClose>
						)}
					</div>
				</SheetFooter>
			</div>
		</SheetContent>
	);
};

export default NodeSidePanel;
