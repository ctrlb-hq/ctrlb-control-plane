import { Changes } from "@/constants/PipelineChangesLog";
import React, { createContext, useContext, useState } from "react";

interface PipelineChangesLogProps {
    changesLog: Changes[];
    addChange: (change: Changes) => void;
    clearChangesLog: () => void;
}

const pipelineLogs = localStorage.getItem("changesLog");


const PipelineChangesLogContext = createContext<PipelineChangesLogProps | undefined>(undefined);

export const PipelineChangesLogProvider = ({ children }: { children: React.ReactNode }) => {
    // Initialize state with data from localStorage
    const [changesLog, setChangesLog] = useState<Changes[]>(pipelineLogs ? JSON.parse(pipelineLogs) : []);

    const addChange = (change: Changes) => {
        // Update the state immediately
        const updatedChangesLog = [...changesLog, change];
        setChangesLog(updatedChangesLog);

        // Persist the updated changesLog to localStorage
        localStorage.setItem("changesLog", JSON.stringify(updatedChangesLog));
    };

    const clearChangesLog = () => {
        localStorage.removeItem("changesLog");
        setChangesLog([]);
    }

    return (
        <PipelineChangesLogContext.Provider value={{ changesLog, addChange, clearChangesLog }}>
            {children}
        </PipelineChangesLogContext.Provider>
    );
};

const usePipelineChangesLog = () => {
    const context = useContext(PipelineChangesLogContext);
    if (!context) {
        throw new Error("usePipelineChangesLog must be used within a PipelineChangesLogProvider");
    }
    return context;
};

export default usePipelineChangesLog;