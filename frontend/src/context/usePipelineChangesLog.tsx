import { Changes } from "@/constants/PipelineChangesLog";
import React, { createContext, useContext, useRef, useState } from "react";

interface PipelineChangesLogProps {
    changesLog: Changes[];
    addChange: (change: Changes) => void;
}

const pipelineLogs = localStorage.getItem("changesLog");

const PipelineChangesLogContext = createContext<PipelineChangesLogProps | undefined>(undefined);

export const PipelineChangesLogProvider = ({ children }: { children: React.ReactNode }) => {
    // Use useRef to store the changesLog for immediate updates
    const changesLogRef = useRef<Changes[]>(pipelineLogs ? JSON.parse(pipelineLogs) : []);

    // Use useState to trigger re-renders when changesLog is updated
    const [changesLog, setChangesLog] = useState<Changes[]>(changesLogRef.current);

    // Function to add a change to the log
    const addChange = (change: Changes) => {
        // Update the ref value
        changesLogRef.current.push(change);

        // Update the state to trigger a re-render
        setChangesLog([...changesLogRef.current]);

        // Persist the updated log to localStorage
        localStorage.setItem("changesLog", JSON.stringify(changesLogRef.current));
    };

    return (
        <PipelineChangesLogContext.Provider value={{ changesLog, addChange }}>
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