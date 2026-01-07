import { useState } from "react";
import { Drawer, IconButton, Button } from "@mui/material";
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

  const nodeChanges = changesLog.filter((change) => change.type !== "Edge");
  const edgeChanges = changesLog.filter((change) => change.type === "Edge");

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
            borderRadius: "10px",
          }}
        >
          <X className="w-5 h-5" />
        </IconButton>
      </div>

      {/* BODY */}
      <div className="flex-1 overflow-y-auto p-4">
        {!isEditFormOpen && (
          <>
            {/* Node Changes Section */}
            {nodeChanges.length > 0 && (
              <div className="mb-6">
                <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase tracking-wide">
                  Node Changes ({nodeChanges.length})
                </h3>
                <div className="space-y-3">
                  {nodeChanges.map((change, index) => (
                    <div
                      key={index}
                      className="flex justify-between items-center p-3 bg-gray-50 rounded-lg hover:bg-gray-100 transition-colors"
                    >
                      <div>
                        <p className="font-medium">{change.type}</p>
                        <p className="text-sm text-gray-600">{change.name}</p>
                      </div>

                      <Edit
                        className="cursor-pointer text-blue-600 hover:text-blue-800 w-5 h-5"
                        onClick={() => openEditForm(change)}
                      />
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Edge Changes Section */}
            {edgeChanges.length > 0 && (
              <div>
                <h3 className="text-sm font-semibold text-gray-700 mb-3 uppercase tracking-wide">
                  Edge Changes ({edgeChanges.length})
                </h3>
                <div className="space-y-3">
                  {edgeChanges.map((change, index) => (
                    <div
                      key={index}
                      className="flex justify-between items-center p-3 bg-blue-50 rounded-lg"
                    >
                      <div>
                        <p className="font-medium">{change.type}</p>
                        <p className="text-sm text-gray-600">{change.name}</p>
                      </div>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* Empty State */}
            {nodeChanges.length === 0 && edgeChanges.length === 0 && (
              <div className="text-center text-gray-500 mt-8">
                <p>No pending changes</p>
              </div>
            )}
          </>
        )}

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