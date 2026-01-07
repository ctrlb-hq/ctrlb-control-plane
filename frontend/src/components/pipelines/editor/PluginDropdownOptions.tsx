import React, { useEffect, useState, useCallback } from "react";
import Button from "@mui/material/Button";
import Menu from "@mui/material/Menu";
import MenuItem from "@mui/material/MenuItem";
import ListSubheader from "@mui/material/ListSubheader";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import { ComponentService } from "@/services/component";
import { JsonSchema } from "@jsonforms/core";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";

interface Plugin {
  name: string;
  display_name: string;
  type: string;
  supported_signals: string[];
}

type AgentType = "otel" | "fluent-bit";

interface Props {
  kind: "receiver" | "processor" | "exporter" | "input" | "filter" | "output";
  nodeType: "source" | "processor" | "destination";
  label: string;
  dataType: "receiver" | "processor" | "exporter" | "input" | "filter" | "output";
  disabled: boolean;
  agentType?: AgentType;
}

const PluginDropdownOptions = React.memo(
  ({
    kind,
    nodeType,
    label,
    dataType,
    disabled,
    agentType = "otel",
  }: Props) => {
    const [anchorEl, setAnchorEl] = useState<null | HTMLElement>(null);
    const menuOpen = Boolean(anchorEl);

    const [isPanelOpen, setIsPanelOpen] = useState(false);
    const [optionValue, setOptionValue] = useState("");
    const [pluginName, setPluginName] = useState<string | undefined>();

    const [plugins, setPlugins] = useState<Plugin[]>([]);
    const [form, setForm] = useState<JsonSchema>({});
    const [config, setConfig] = useState<object>({});
    const [uiSchema, setUiSchema] = useState({
      type: "VerticalLayout",
      elements: [],
    });

    const { addNode } = useGraphFlow();

    const fetchPlugins = useCallback(async () => {
      const res = await ComponentService.getTransporterService(kind);
      setPlugins(res || []);
    }, [kind]);

    useEffect(() => {
      fetchPlugins();
    }, [fetchPlugins, agentType]);

    const fetchForm = async (plugin: string) => {
      const schema = await ComponentService.getTransporterForm(plugin);
      const ui = await ComponentService.getTransporterUiSchema(plugin);
      setForm(schema);
      setUiSchema(ui);
    };

    const handleMenuOpen = (event: React.MouseEvent<HTMLButtonElement>) => {
      setAnchorEl(event.currentTarget);
    };

    const handleMenuClose = () => {
      setAnchorEl(null);
    };

    const handleSelectPlugin = (plugin: Plugin) => {
      setPluginName(plugin.name);
      setOptionValue(plugin.display_name);
      setConfig({});
      setForm({});
      setIsPanelOpen(true);
      handleMenuClose();
      fetchForm(plugin.name);
    };

    const handleSubmit = (submittedConfig: any) => {
      const supported_signals = plugins.find(
        (p) => p.name === pluginName,
      )?.supported_signals;

      addNode({
        type: nodeType,
        position: { x: 0, y: 0 },
        data: {
          type: dataType,
          name: optionValue,
          supported_signals,
          component_name: pluginName,
          config: submittedConfig,
        },
      });

      setIsPanelOpen(false);
    };


    return (
      <>
        <Button
          variant="outlined"
          disabled={disabled}
          onClick={handleMenuOpen}
          sx={{
            display: "flex",
            alignItems: "center",
            gap: 1,
            px: 2,
            py: 1,
            borderWidth: 2,
            borderRadius: "6px",
            textTransform: "none",
            boxShadow: "0 2px 6px rgba(0,0,0,0.08)",
            fontSize: "13px",
            "&:hover": {
              backgroundColor: disabled ? undefined : "#F3F4F6",
            },
          }}
        >
          ➕ Add {label}
        </Button>
        <Menu
			anchorEl={anchorEl}
			open={menuOpen}
			onClose={handleMenuClose}
			anchorOrigin={{
				vertical: "top",
				horizontal: "center",
			}}
			transformOrigin={{
				vertical: "bottom",
				horizontal: "center",
			}}
			PaperProps={{
				sx: {
				width: 300,
				maxHeight: 360,
				borderRadius: "8px",
				boxShadow: "0 8px 24px rgba(0,0,0,0.12)",
				},
			}}
		>
          <ListSubheader
            sx={{
              fontSize: "13px",
              fontWeight: 600,
              lineHeight: 1.4,
              color: "#374151",
              backgroundColor: "#fff",
            }}
          >
            Select a {label}
          </ListSubheader>

          {plugins.map((plugin) => (
            <MenuItem
              key={plugin.name}
              onClick={() => handleSelectPlugin(plugin)}
              sx={{
                fontSize: "13px",
                lineHeight: 1.4,
                whiteSpace: "normal",
                wordBreak: "break-word",
                px: 1.5,
                py: 0.75,
                "&:hover": {
                  backgroundColor: "#F3F4F6",
                },
              }}
            >
              {plugin.display_name}
            </MenuItem>
          ))}
        </Menu>
        <NodeSidePanel
          title={optionValue}
          formSchema={form}
          uiSchema={uiSchema}
          config={config}
          setConfig={setConfig}
          submitLabel={`Add ${label}`}
          submitDisabled
          onSubmit={handleSubmit}
          onDiscard={() => setIsPanelOpen(false)}
          showDelete={false}
          isOpen={isPanelOpen}
        />
      </>
    );
  },
);

export default PluginDropdownOptions;
