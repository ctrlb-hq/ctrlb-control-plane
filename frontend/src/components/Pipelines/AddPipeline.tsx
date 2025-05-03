import { Button } from '../ui/button'
import { PlusIcon } from 'lucide-react'
import {
    Sheet,
    SheetContent,
    SheetDescription,
    SheetHeader,
    SheetTitle,
    SheetTrigger,
} from "@/components/ui/sheet"

import {
    Dialog,
    DialogContent,
    DialogHeader,
    DialogFooter,
    DialogTitle,
    DialogDescription,
} from "@/components/ui/dialog";

import PipelineDetails from './AddNewPipelineComponent/PipelineDetails';
import { usePipelineStatus } from '@/context/usePipelineStatus';
import { useState } from 'react';
import PipelineCanvas from './AddNewPipelineComponent/PipelineCanvas';
import { NodeValueProvider } from '@/context/useNodeContext';


const LandingView = () => {
    const pipelineStatus = usePipelineStatus();
    if (!pipelineStatus) {
        return null;
    }
    const { currentStep, setCurrentStep } = pipelineStatus;
    const [isSheetOpen, setIsSheetOpen] = useState(false);
    const [isDialogOpen, setIsDialogOpen] = useState(false);

    const handleDialogOkay = () => {
        localStorage.removeItem('Sources');
        localStorage.removeItem('Destination');
        localStorage.removeItem('pipelinename');
        localStorage.removeItem("selectedAgentIds");
        localStorage.removeItem("PipelineEdges");
        localStorage.removeItem("Nodes");
        localStorage.removeItem("platform")

        setCurrentStep(0);
        setIsDialogOpen(false);
        setIsSheetOpen(false);
    };

    const handleDialogCancel = () => {
        setIsDialogOpen(false);
    };


    return (
        <div className="flex flex-col gap-7 justify-center items-center">
            <Sheet
                open={isSheetOpen}
                onOpenChange={(open) => {
                    if (!open) {
                        setIsDialogOpen(true);
                    } else {
                        setIsSheetOpen(true);
                    }
                }}
            >
                <SheetTrigger asChild>
                    <Button className="flex gap-1 px-4 py-1 bg-blue-500 text-white" variant="outline">Add New Pipeline
                        <PlusIcon className="h-4 w-4" />
                    </Button>
                </SheetTrigger>
                <SheetContent>
                    {
                        currentStep == 0 ? <PipelineDetails /> : <NodeValueProvider><PipelineCanvas /></NodeValueProvider>
                    }
                </SheetContent>
            </Sheet>
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
                <DialogContent className="w-[50rem]">
                    <DialogHeader>
                        <DialogTitle>Discard Changes?</DialogTitle>
                        <DialogDescription>
                            Are you sure you want to discard the current pipeline setup? 
                        </DialogDescription>
                    </DialogHeader>
                    <DialogFooter>
                        <Button variant="outline" onClick={handleDialogCancel}>
                            Cancel
                        </Button>
                        <Button className="bg-blue-500" onClick={handleDialogOkay}>
                            OK
                        </Button>
                    </DialogFooter>
                </DialogContent>
            </Dialog>
        </div>
    )
}

export default LandingView



