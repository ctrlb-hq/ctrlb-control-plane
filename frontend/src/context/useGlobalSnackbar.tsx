import React, { createContext, useContext, useState } from "react";
import { Snackbar, Alert, AlertColor } from "@mui/material";
import { Loader2 } from "lucide-react";

type SnackbarVariant = AlertColor | "loading";

type SnackbarContextType = {
  showSnackbar: (
    message: string,
    variant?: SnackbarVariant,
    duration?: number
  ) => void;
};

const SnackbarContext = createContext<SnackbarContextType | null>(null);

export const GlobalSnackbarProvider: React.FC<{
  children: React.ReactNode;
}> = ({ children }) => {
  const [open, setOpen] = useState(false);
  const [message, setMessage] = useState("");
  const [variant, setVariant] = useState<SnackbarVariant>("success");
  const [duration, setDuration] = useState<number | null>(4000);

  const showSnackbar = (
    message: string,
    variant: SnackbarVariant = "success",
    duration: number = 4000
  ) => {
    setMessage(message);
    setVariant(variant);

    if (variant === "loading") {
      setDuration(null); 
    } else {
      setDuration(duration);
    }

    setOpen(true);
  };

  const handleClose = (
    _: React.SyntheticEvent | Event,
    reason?: string
  ) => {
    if (variant === "loading") return;
    if (reason === "clickaway") return;
    setOpen(false);
  };

  return (
    <SnackbarContext.Provider value={{ showSnackbar }}>
      {children}

      <Snackbar
        open={open}
        autoHideDuration={duration ?? undefined}
        onClose={handleClose}
        anchorOrigin={{ vertical: "bottom", horizontal: "right" }}
      >
        <Alert
          severity={variant === "loading" ? "info" : variant}
          variant="filled"
          icon={
            variant === "loading" ? (
              <Loader2 className="h-4 w-4 animate-spin" />
            ) : undefined
          }
          onClose={variant === "loading" ? undefined : handleClose}
          sx={{ width: "100%" }}
        >
          {message}
        </Alert>
      </Snackbar>
    </SnackbarContext.Provider>
  );
};


export const useGlobalSnackbar = () => {
  const ctx = useContext(SnackbarContext);
  if (!ctx) {
    throw new Error(
      "useGlobalSnackbar must be used within GlobalSnackbarProvider"
    );
  }
  return ctx;
};
