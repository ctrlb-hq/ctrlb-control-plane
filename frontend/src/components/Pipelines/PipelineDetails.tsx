import { useEffect } from "react";
import { Boxes, Edit, Trash2 } from "lucide-react";
import { useRef, useState, useCallback, useMemo } from "react";
import EditPipelineYAML from "./EditPipelineYAML";
import ReactFlow, {
    MiniMap,
    Controls,
    Background,
    useEdgesState,
    addEdge,
    Node,
    Edge,
    Connection,
    ReactFlowInstance,
    EdgeMouseHandler,
    Panel,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { Sheet, SheetContent, SheetTitle, SheetTrigger, SheetDescription, SheetClose } from '@/components/ui/sheet';
import { Button } from '@/components/ui/button';
import { SourceNode } from "../CanvasForPipelines/SourceNode";
import { ProcessorNode } from "../CanvasForPipelines/ProcessorNode";
import { DestinationNode } from "../CanvasForPipelines/DestinationNode";
import { Switch } from "../ui/switch";
import {
    Dialog,
    DialogClose,
    DialogContent,
    DialogDescription,
    DialogFooter,
    DialogHeader,
    DialogTitle,
    DialogTrigger,
} from "@/components/ui/dialog"
import { initialEdges } from "@/constants/PipelineNodeAndEdges";
import { Label } from "../ui/label";
import SourceDropdownOptions from "./DropdownOptions/SourceDropdownOptions";
import { useNodeValue } from "@/context/useNodeContext";
import DestinationDropdownOptions from "./DropdownOptions/DestinationDropdownOptions";
import ProcessorDropdownOptions from "./DropdownOptions/ProcessorDropdownOptions";
import usePipelineChangesLog from "@/context/usePipelineChangesLog";
import { useToast } from "@/hooks/use-toast";
import pipelineServices from "@/services/pipelineServices";
import { Pipeline } from "@/types/pipeline.types";
import { Agents } from "@/types/agent.types";


const PipelineDetails = ({ pipelineId }: { pipelineId: string }) => {
    const TABS = [
        { label: "Overview", value: "overview" },
        { label: "YAML", value: "yaml" },
    ];
    const [agentValues, setAgentValues] = useState<Agents[]>([])
    const { nodeValue, setNodeValue, onNodesChange } = useNodeValue();
    const [edges, setEdges, onEdgesChange] = useEdgesState(initialEdges);
    const reactFlowWrapper = useRef<HTMLDivElement>(null);
    const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
    const [isEditMode, setIsEditMode] = useState(false);
    const [selectedEdge, setSelectedEdge] = useState<Edge | null>(null);
    const [edgePopoverPosition, setEdgePopoverPosition] = useState({ x: 0, y: 0 });
    const { changesLog } = usePipelineChangesLog()
    const [pipelineOverview, setPipelineOverview] = useState<Pipeline>()
    const [isOpen, setIsOpen] = useState(false)
    const { toast } = useToast()


    const nodeTypes = useMemo(() => ({
        source: SourceNode,
        processor: ProcessorNode,
        destination: DestinationNode
    }), [])

    const formatTimestamp = (timestamp: number | undefined) => {
        if (!timestamp) return "N/A";
        const date = new Date(timestamp * 1000); // Convert seconds to milliseconds
        const hours = date.getHours().toString().padStart(2, '0')
        const minutes = date.getMinutes().toString().padStart(2, '0')
        return `${hours}:${minutes}`
    }

    const createdBy = localStorage.getItem("userEmail")

    const handleGetPipeline = async () => {
        const res = await pipelineServices.getPipelineById(pipelineId)
        setPipelineOverview(res)
    }

    const handleGetConnectedAgentsToPipeline = async () => {
        const res = await pipelineServices.getAllAgentsAttachedToPipeline(pipelineId)
        setAgentValues(res)
    }


    const handleGetPipelineGraph = async () => {
        const res = await pipelineServices.getPipelineGraph(pipelineId);
        const edges=res.edges
        const updatedNodes = res.nodes.map((node: any) => {
            const nodeType = node.component_role === 'receiver' ? 'source' : node.component_role === 'exporter' ? 'destination' : 'processor';
            return {
                id: node.component_id.toString(),
                type: nodeType,
                position: { x: Math.random() * 80, y: Math.random() * 60 },
                data: {
                    id: node.component_id.toString(),
                    name: node.name,
                    component_name: node.component_name,
                    supported_signals: node.supported_signals,
                    config: node.config,
                },
            };
        });
        setNodeValue(updatedNodes);
        setEdges(edges)
    };

    useEffect(() => {
        handleGetPipelineGraph()
    }, [pipelineId])



    useEffect(() => {
        handleGetPipeline()
        handleGetConnectedAgentsToPipeline()
    }, [isEditMode]);

    const onConnect = useCallback(
        (params: Edge | Connection) => setEdges((eds) => addEdge({ ...params, animated: true }, eds)),
        [setEdges],
    );


    const onEdgeClick: EdgeMouseHandler = useCallback((event, edge) => {
        if (!isEditMode) return;
        // Calculate the position for the popover
        const rect = reactFlowWrapper.current?.getBoundingClientRect();
        if (rect) {
            setEdgePopoverPosition({
                x: event.clientX - rect.left,
                y: event.clientY - rect.top,
            });
        }

        setSelectedEdge(edge);
    }, [isEditMode]);

    const handleDeleteEdge = useCallback(() => {
        if (selectedEdge) {
            setEdges((edges) => edges.filter((edge) => edge.id !== selectedEdge.id));
            setSelectedEdge(null);
        }
    }, [selectedEdge, setEdges]);

    const [activeTab, setActiveTab] = useState("overview");

    // Close popover when clicking elsewhere
    const onPaneClick = useCallback(() => {
        setSelectedEdge(null);
    }, []);

    //validation of YAML files and the output given will be shown in the toast
    //error or success
    const handleDeployChanges = () => {
        setTimeout(() => {
            toast({
                title: "Success",
                description: "Successfully deployed the pipeline",
                duration: 3000,
            });
        }, 2000);
    }

    const handleDeletePipeline = async () => {
        await pipelineServices.deletePipelineById(pipelineId);
        setIsOpen(false);
        window.location.reload();
    }

    return (
        <div className="py-4 flex flex-col">
            <div className="flex mb-5 gap-2 items-center">
                <Boxes className="text-gray-700" size={36} />
                <h1 className="text-2xl text-gray-800">{pipelineOverview?.name}</h1>
            </div>
            <div className="flex items-center w-full md:w-auto">
                <div className="flex gap-2 justify-between w-full mb-2">
                    <div className="flex gap-2 justify-start">
                        {TABS.map(({ label, value }) => (
                            <button
                                key={value}
                                onClick={() => setActiveTab(value)}
                                className={`px-4 py-2 text-lg rounded-t-md text-gray-600 focus:outline-none ${activeTab === value
                                    ? "border-b-2 border-blue-500 text-blue-500 font-semibold"
                                    : ""
                                    }`}
                            >
                                {label}
                            </button>
                        ))}
                    </div>
                    <div className="flex gap-2">
                        <Sheet>
                            <SheetTrigger asChild>
                                <Button className="bg-blue-500">View/Edit Pipeline</Button>
                            </SheetTrigger>
                            <SheetContent className="w-full sm:max-w-full p-0" side="right">
                                <div className="flex justify-between items-center p-4 border-b">
                                    <div className="flex items-center space-x-2">
                                        <div className="text-xl font-medium">ctrlb</div>
                                    </div>
                                    <div className="flex items-center mx-4">
                                        <Sheet>
                                            <SheetTrigger asChild>
                                                <Button className="rounded-full px-6">Review</Button>
                                            </SheetTrigger>
                                            <SheetContent className="w-[30rem]">
                                                <SheetTitle>Pending Changes</SheetTitle>
                                                <SheetDescription>
                                                    <div className="flex flex-col gap-6 mt-4 overflow-auto h-[40rem]">
                                                        {
                                                            changesLog.map((change, index) => (
                                                                <div key={index} className="flex justify-between items-center">
                                                                    <div className="flex flex-col">
                                                                        <p className="text-lg">{change.type}</p>
                                                                        <p className="text-lg text-gray-800">{change.name}</p>
                                                                    </div>
                                                                    <div className="flex justify-end gap-3 items-center">
                                                                        <p className={`${change.status == 'added' ? "text-green-500" : change.status == 'deleted' ? "text-red-500" : "text-gray-600"} text-lg`}>[{change.status}]</p>
                                                                        <Edit size={20} />
                                                                    </div>
                                                                </div>
                                                            ))
                                                        }
                                                    </div>
                                                </SheetDescription>
                                                <SheetClose className="flex justify-end mt-4 w-full">
                                                    <div>
                                                        <Button onClick={handleDeployChanges} className="bg-blue-500">Deploy Changes</Button>
                                                    </div>
                                                </SheetClose>

                                            </SheetContent>

                                        </Sheet>
                                        <div className="mx-4 flex items-center space-x-2">
                                            <Switch id="edit-mode" checked={isEditMode} onCheckedChange={setIsEditMode} />
                                            <Label htmlFor="edit-mode">Edit Mode</Label>
                                        </div>
                                    </div>
                                </div>
                                <div style={{ height: '77vh', backgroundColor: "#f9f9f9" }} ref={reactFlowWrapper}>
                                    <ReactFlow

                                        nodes={nodeValue}
                                        edges={edges.map(edge => ({
                                            ...edge,
                                            animated:true,
                                            label: isEditMode ? '' : edge.label
                                        }))}
                                        onNodesChange={onNodesChange}
                                        onEdgesChange={onEdgesChange}
                                        onConnect={onConnect}
                                        nodeTypes={nodeTypes}
                                        onInit={setReactFlowInstance}
                                        onEdgeClick={onEdgeClick}
                                        onPaneClick={onPaneClick}
                                        fitView
                                    >
                                        <Background />
                                        <Controls />
                                        <MiniMap />
                                        {selectedEdge && isEditMode && (
                                            <Panel
                                                position="top-left"
                                                style={{
                                                    position: 'absolute',
                                                    left: edgePopoverPosition.x,
                                                    top: edgePopoverPosition.y,
                                                    transform: 'translate(-50%, -50%)',
                                                    background: 'white',
                                                    padding: '8px',
                                                    borderRadius: '4px',
                                                    boxShadow: '0 2px 4px rgba(0,0,0,0.2)',
                                                    zIndex: 10
                                                }}
                                            >
                                                <Trash2 onClick={handleDeleteEdge} className="text-red-500 cursor-pointer" size={16} />
                                            </Panel>
                                        )}
                                    </ReactFlow>
                                </div>

                                <div className="bg-gray-100 h-1/5 p-4 rounded-lg">
                                    <div className="flex justify-around gap-2">
                                        <div className='flex items-center'>
                                            <SourceDropdownOptions />
                                        </div>
                                        <div className='flex items-center'>
                                            <ProcessorDropdownOptions />
                                        </div>

                                        <div className='flex items-center'>
                                            <DestinationDropdownOptions />
                                        </div>
                                    </div>
                                </div>
                            </SheetContent>
                        </Sheet>
                        <Dialog open={isOpen} onOpenChange={setIsOpen}>
                            <DialogTrigger asChild>
                                <Button variant="destructive">Delete Pipeline</Button>
                            </DialogTrigger>
                            <DialogContent className="sm:max-w-[40rem] h-[25rem]">
                                <DialogHeader>
                                    <DialogTitle className="text-red-500 text-xl">Delete Pipeline</DialogTitle>
                                    <DialogDescription className="text-md text-gray-700">
                                        Are you sure you want to delete this Pipeline?
                                    </DialogDescription>
                                </DialogHeader>
                                <div className="flex flex-col">
                                    <p className="text-gray-600">Pipeline Id: {pipelineOverview?.id} </p>
                                    <p className="text-gray-600">Pipeline Name: {pipelineOverview?.name}</p>
                                    <p className="text-red-500 mt-2">After Deleting this pipeline the below agents will be orphaned</p>
                                    {agentValues && agentValues.map((agent, index) => (
                                        <p className="text-gray-600" key={index}>
                                            Agent: {agent.name}
                                        </p>
                                    ))}
                                </div>
                                <DialogFooter>
                                    <DialogClose asChild>
                                        <Button>Cancel</Button>
                                    </DialogClose>
                                    <Button onClick={handleDeletePipeline} variant={"destructive"} >Delete</Button>
                                </DialogFooter>
                            </DialogContent>
                        </Dialog>
                    </div>
                </div>
            </div>
            {activeTab == "overview" ? <div className="flex flex-col w-[30rem] md:w-full">
                <div className="flex flex-col py-2">
                    <p className="capitalize">Name: {pipelineOverview?.name}</p>
                    <p>Created By: {createdBy}</p>
                    <p>Created At: {formatTimestamp(pipelineOverview?.created_at)}</p>
                    <p>Updated At: {formatTimestamp(pipelineOverview?.updated_at)}</p>
                </div>
            </div> : <EditPipelineYAML />}

        </div>
    )
}

export default PipelineDetails