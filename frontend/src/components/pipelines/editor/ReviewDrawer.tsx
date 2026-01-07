import { useState } from "react";
import { Drawer, IconButton, Button }from "@mui/material";
import { Edit, Loader2, X } from "lucide-react";
import NodeSidePanel from "@/components/pipelines/editor/NodeSidePanel";
import { ComponentService } from "@/services/component";

interface ReviewDrawerProps {
  open: boolean;
  changesLog: any[];
  isDeploying: boolean;

  onClose: () => void;
  onDeploy: () => void;
  onApplyConfig: (id: string, config: any) => void;
}

const ReviewDrawer: React.FC<ReviewDrawerProps> = ({
  open,
  changesLog,
  isDeploying,
  onClose,
  onDeploy,
  onApplyConfig,
}) => {
  const [isEditFormOpen, setIsEditFormOpen] = useState(false);
  const [selectedChange, setSelectedChange] = useState<any>(null);
  const [form, setForm] = useState<any>({});
  const [uiSchema, setUiSchema] = useState<any>({});
  const [config, setConfig] = useState<any>({});

  const openEditForm = async (change: any) => {
    setSelectedChange(change);
    setIsEditFormOpen(true);

    setForm(await ComponentService.getTransporterForm(change.component_type));
    setUiSchema(
      await ComponentService.getTransporterUiSchema(change.component_type),
    );
    setConfig(change.finalConfig);
  };

  const handleApply = () => {
    if (!selectedChange) return;

    onApplyConfig(selectedChange.id, config);

    setIsEditFormOpen(false);
    setSelectedChange(null);
  };

  const handleClose = () => {
    setIsEditFormOpen(false);
    setSelectedChange(null);
    onClose();
  };

  return (
    <Drawer
      anchor="right"
      open={open}
      onClose={handleClose}
      PaperProps={{
        sx: { width: 480, display: "flex", flexDirection: "column" },
      }}
    >
      {/* HEADER */}
      <div className="flex justify-between items-center p-4 border-b">
        <h2 className="font-semibold">
          {isEditFormOpen ? "Edit Configuration" : "Pending Changes"}
        </h2>

        <IconButton
          onClick={handleClose}
          sx={{
            "&:hover": {
              backgroundColor: "#EF4444",
              color: "#fff",
            },
            borderRadius:"10px"
          }}
        >
          <X className="w-5 h-5" />
        </IconButton>
      </div>

      {/* BODY */}
      <div className="flex-1 overflow-y-auto p-4">
        {!isEditFormOpen &&
          changesLog.map((change, index) => (
            <div
              key={index}
              className="flex justify-between items-center mb-4"
            >
              <div>
                <p className="font-medium">{change.type}</p>
                <p className="text-gray-600">{change.name}</p>
              </div>

              {change.type !== "Edge" && (
                <Edit
                  className="cursor-pointer"
                  onClick={() => openEditForm(change)}
                />
              )}
            </div>
          ))}

        {isEditFormOpen && selectedChange && (
          <NodeSidePanel
            title={selectedChange.name}
            formSchema={form}
            uiSchema={uiSchema}
            config={config}
            setConfig={setConfig}
            submitLabel="Apply"
            onSubmit={handleApply}
            onDiscard={handleClose}
            showDelete={false}
            isOpen
          />
        )}
      </div>

      {/* FOOTER */}
      {!isEditFormOpen && (
        <div className="p-4 border-t">
          <Button
            variant="contained"
            onClick={onDeploy}
            color="primary"
            disabled={isDeploying}
            className="bg-blue-500 flex items-center gap-2"
          >
            {isDeploying ? (
              <>
                <Loader2 className="h-4 w-4 animate-spin" />
                Deploying…
              </>
            ) : (
              "Deploy Changes"
            )}
          </Button>
        </div>
      )}
    </Drawer>
  );
};

export default ReviewDrawer;
