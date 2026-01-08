import { Dispatch, SetStateAction, useState } from "react";
import {
  Drawer,
  Box,
  Dialog,
  DialogTitle,
  DialogActions,
  DialogContent,
  DialogContentText,
  Button,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import AddPipelineDetails from "@/components/pipelines/create/AddPipelineDetails";
import { useGraphFlow } from "@/context/useGraphFlowContext";

interface AddPipelineSheetProps {
  isOpen: boolean;
  setIsOpen: Dispatch<SetStateAction<boolean>>;
}

const AddPipelineSheet = ({ isOpen, setIsOpen }: AddPipelineSheetProps) => {
  const [isDialogOpen, setIsDialogOpen] = useState(false);

  const { resetGraph, changesLog } = useGraphFlow();

  const shouldShowDialog = () => {
    return changesLog.length > 0;
  };

  const handleDrawerClose = () => {
    if (shouldShowDialog()) {
      setIsDialogOpen(true);
    } else {
      setIsOpen(false);
    }
  };

  const handleDialogOkay = () => {
    localStorage.removeItem("Sources");
    localStorage.removeItem("Destination");
    localStorage.removeItem("pipelinename");
    localStorage.removeItem("selectedAgentIds");
    localStorage.removeItem("changesLog");
    localStorage.removeItem("platform");
    localStorage.removeItem("agentType");

    resetGraph();
    setIsDialogOpen(false);
    setIsOpen(false);
  };

  const handleDialogCancel = () => {
    setIsDialogOpen(false);
  };

  return (
    <>
      <Drawer
        anchor="right"
        open={isOpen}
        onClose={handleDrawerClose}
        PaperProps={{
          sx: { width: "70vw" },
        }}
      >
        <Button
          onClick={handleDrawerClose}
          variant="text"
          sx={{
            position: "absolute",
            top: 8,
            right: 8,
            minWidth: "auto",
            padding: "3px",
            color: "black",
            zIndex: 10,
            borderRadius: "10px",
          }}
          className="hover:bg-red-500 hover:text-white"
        >
          <CloseIcon />
        </Button>

        <Box sx={{ height: "100%", overflow: "auto", pt: 4 }}>
          <AddPipelineDetails />
        </Box>
      </Drawer>

      <Dialog
        open={isDialogOpen}
        onClose={handleDialogCancel}
        aria-labelledby="discard-dialog-title"
        aria-describedby="discard-dialog-description"
      >
        <DialogTitle id="discard-dialog-title">
          Discard Pipeline?
        </DialogTitle>

        <DialogContent>
          <DialogContentText id="discard-dialog-description">
            Your changes will be lost. Continue?
          </DialogContentText>
        </DialogContent>

        <DialogActions>
          <Button variant="contained" color="error" onClick={handleDialogCancel}>
            Cancel
          </Button>
          <Button variant="contained" color="primary" onClick={handleDialogOkay}>
            OK
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};

export default AddPipelineSheet;
