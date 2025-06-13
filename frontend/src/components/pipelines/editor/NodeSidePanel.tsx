import { createTheme, ThemeProvider } from "@mui/material/styles";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { Button } from "@/components/ui/button";
import { SheetFooter, SheetClose, SheetContent } from "@/components/ui/sheet";
import { ArrowBigRightDash } from "lucide-react";
import { customEnumRenderer } from "@/components/pipelines/editor/CustomEnumControl";

interface NodeSidePanel {
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

const NodeSidePanel: React.FC<NodeSidePanel> = ({
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
						marginBottom: "0.5rem",
					},
				},
			},
		},
	});

	const renderers = [...materialRenderers, customEnumRenderer];
	return (
		<SheetContent className="w-[36rem]">
			<div className="flex flex-col gap-4 p-4">
				<div className="flex gap-3 items-center">
					<ArrowBigRightDash className="w-6 h-6" />
					<h2 className="text-xl font-bold">{title}</h2>
				</div>
				{description && (
					<p className="text-gray-500">
						{description} <span className="text-blue-500 underline">Documentation</span>
					</p>
				)}
				<ThemeProvider theme={theme}>
					<div className="mt-3">
						<div className="p-3">
							<div className="overflow-y-auto h-[32rem] pt-3">
								{formSchema && (
									<JsonForms
										data={config}
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
								)}
							</div>
						</div>
					</div>
				</ThemeProvider>
				<SheetFooter>
					<SheetClose>
						<div className="flex gap-3">
							<Button className="bg-blue-500" onClick={onSubmit} disabled={submitDisabled}>
								{submitLabel}
							</Button>
							<Button variant={"outline"} onClick={onDiscard}>
								Discard Changes
							</Button>
							{showDelete && (
								<Button variant={"outline"} onClick={onDelete}>
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
