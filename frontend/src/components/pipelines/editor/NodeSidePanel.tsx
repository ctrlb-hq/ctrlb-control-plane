import { useEffect, useMemo, useState, useRef } from "react";
import Ajv from "ajv";
import  { Drawer, IconButton, Button, createTheme, ThemeProvider } from "@mui/material";
import { JsonForms } from "@jsonforms/react";
import { materialCells, materialRenderers } from "@jsonforms/material-renderers";
import { createAjv } from "@jsonforms/core";
import { ArrowBigRightDash, X } from "lucide-react";
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
  onSubmit: (data: any) => void;
  onDiscard: () => void;
  onDelete?: () => void;
  showDelete?: boolean;
  isOpen?: boolean;
}

const ajv = new Ajv({ useDefaults: true, allErrors: true, strict: false });

const applySchemaDefaults = (schema: any, data: any) => {
  const clonedData = JSON.parse(JSON.stringify(data || {}));
  const validateWithDefaults = ajv.compile(schema);
  validateWithDefaults(clonedData);
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
  isOpen = false,
}) => {
  const [showErrors, setShowErrors] = useState(false);
  const [draftConfig, setDraftConfig] = useState(() =>
    applySchemaDefaults(formSchema, config),
  );
  const [formErrors, setFormErrors] = useState<any[]>([]);

  const lastConfigRef = useRef<string>("");

  useEffect(() => {
    const newConfigString = JSON.stringify(config);
    if (newConfigString === lastConfigRef.current) return;

    lastConfigRef.current = newConfigString;
    setDraftConfig(applySchemaDefaults(formSchema, config));
    setShowErrors(false);
  }, [formSchema, config]);

  const defaultsAjv = useMemo(
    () => createAjv({ useDefaults: true, allErrors: true, strict: false }),
    [],
  );

  const validate = useMemo(() => ajv.compile(formSchema), [formSchema]);

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
                backgroundColor: "#3B82F6",
                color: "#FFFFFF",
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
    onSubmit(draftConfig);
  };

  const handleDiscard = () => {
    setShowErrors(false);
    setDraftConfig(applySchemaDefaults(formSchema, config));
    onDiscard();
  };

  return (
    <Drawer
      anchor="right"
      open={isOpen}
      onClose={handleDiscard}
      PaperProps={{
        sx: {
          width: 550,
          display: "flex",
          flexDirection: "column",
        },
      }}
    >
      <div className="flex items-center justify-between p-4 border-b">
        <div className="flex gap-3 items-center">
          <ArrowBigRightDash className="w-6 h-6" />
          <h2 className="text-xl font-semibold">{title}</h2>
        </div>
        <IconButton
          onClick={handleDiscard}
          aria-label="Close"
          sx={{
            width: 32,
            height: 32,
            color: "#6B7280",
            transition: "all 0.2s ease",
            "&:hover": {
              backgroundColor: "#EF4444", 
              color: "#FFFFFF",
            },
            borderRadius:"10px"
          }}
        >
          <X className="w-5 h-5" />
        </IconButton>
      </div>

      <div className="px-4 pt-3 text-sm text-gray-500">
        For more information please refer{" "}
        <span className="text-blue-500 underline cursor-pointer">
          Documentation
        </span>
      </div>

      <ThemeProvider theme={theme}>
        <div className="flex-grow overflow-y-auto px-4 pt-4 text-xs">
          <JsonForms
            data={draftConfig}
            schema={formSchema}
            uischema={uiSchema}
            renderers={renderers}
            cells={materialCells}
            ajv={defaultsAjv}
            validationMode={showErrors ? "ValidateAndShow" : "ValidateAndHide"}
            onChange={({ data }) => {
              setDraftConfig(data);
              const valid = validate(data);
              setFormErrors(valid ? [] : validate.errors || []);
            }}
          />
        </div>
      </ThemeProvider>

      <div className="p-4 border-t flex gap-3">
        <Button
          variant="contained"
          color="primary"
          className=" text-sm"
          onClick={handleSubmit}
          disabled={submitDisabled && formErrors.length > 0}
        >
          {submitLabel}
        </Button>

        <Button variant="outlined" className="text-sm" color="error" onClick={handleDiscard}>
          Discard Changes
        </Button>

        {showDelete && (
          <Button
            variant="contained"
            className="text-sm"
            color="error"
            onClick={onDelete}
          >
            Delete Node
          </Button>
        )}
      </div>
    </Drawer>
  );
};

export default NodeSidePanel;
