import { TableCell } from "@/components/ui/table"
import { RefreshCcwIcon } from "lucide-react";
import { Button } from "../ui/button";
import { usePipelineOverview } from "@/context/usePipelineDetailContext";
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
import {
    Select,
    SelectContent,
    SelectGroup,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"
import { useEffect, useState } from "react";
import pipelineServices from "@/services/pipelineServices";
import { Agents } from "@/types/agent.types";
import { useAgentValues } from "@/context/useAgentsValues";


const PipelineOverviewTable = ({ pipelineId }: { pipelineId: string }) => {
    const { pipelineOverview } = usePipelineOverview()
    const [selectedAgent, setSelectedAgent] = useState<Agents | null>(null);
    const [totalAgent, setTotalAgent] = useState<Agents[]>([])
    const [connectedAgent, setConnectedAgent] = useState<Agents[]>([])
    const { agentValues, setAgentValues } = useAgentValues()

    const handleSelectDevice = (id: string) => {
        setAgentValues(agentValues.map(device =>
            device.id === id ? { ...device, selected: !device.selected } : device
        ));
    };

    const getAgentsConnectToPipeline = async () => {
        const authToken = localStorage.getItem('authToken');
        if (!authToken) {
            console.error("Unauthorized: No authToken found. Skipping agent fetch.");
            return;
        }
        try {
            const res = await pipelineServices.getAllAgentsAttachedToPipeline(pipelineId);
            setConnectedAgent(res);
            setTotalAgent(agentValues);
        } catch (error) {
            console.error("Failed to fetch agents:", error);
        }
    };

    useEffect(() => {
        if (localStorage.getItem('authToken'))
            getAgentsConnectToPipeline()
    }, [connectedAgent,setConnectedAgent])

    const handleAgentApply = async (agent: Agents) => {
        console.log("agent is: ", agent)
        await pipelineServices.attachAgentToPipeline(pipelineId, agent.id)
        setConnectedAgent([...connectedAgent, {
            id: selectedAgent?.id!,
            name: selectedAgent?.name!,
            status: "unknown",
            pipeline_name: pipelineOverview?.name || "",
            version: selectedAgent?.version!,
            log_rate: selectedAgent?.log_rate!,
            metrics_rate: selectedAgent?.metrics_rate!,
            trace_rate: selectedAgent?.trace_rate!,
            selected: false
        }]);
        getAgentsConnectToPipeline()
    }

    const handleDetachAgent = async (ids: string[]) => {
        if (ids.length > 1) {
            for (const id of ids) {
                const res = await pipelineServices.detachAgentFromPipeline(pipelineId, id);
                console.log(`Detached agent with id: ${id}`, res);
            }
        } else if (ids.length === 1) {
            const res = await pipelineServices.detachAgentFromPipeline(pipelineId, ids[0]);
            console.log(`Detached agent with id: ${ids[0]}`, res);
        } else {
            console.log("No agents to detach");
        }

        // Refresh the table UI by fetching the updated agents
        await getAgentsConnectToPipeline();
    }

    return (
        <div className="p-4 rounded-lg shadow">
            <div className="flex mb-5 justify-between">
                <h1 className="text-xl flex justify-center items-center text-gray-600">Agents
                    <RefreshCcwIcon className="w-5 mx-4 text-blue-500" />
                </h1>
                {agentValues && agentValues.every(agent => agent.selected) ? (
                    <Dialog>
                        <DialogTrigger>
                            <Button variant={"destructive"}>Detach All Agents</Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[425px] h-[16rem]">
                            <DialogHeader>
                                <DialogTitle className="mb-2">Detach All Agents</DialogTitle>
                                <DialogDescription>
                                    <p className="text-gray-700 mb-4">Are you sure you want to detach all agents from {pipelineOverview?.name} Pipeline?</p>
                                    {agentValues.filter(agent => agent.selected).map(
                                        agent => (
                                            <p className="text-gray-600" key={agent.id}>{agent.name}</p>
                                        )
                                    )}
                                </DialogDescription>
                            </DialogHeader>
                            <DialogFooter>
                                <DialogClose className="flex gap-4">
                                    <Button>Cancel</Button>
                                    <Button onClick={() => { handleDetachAgent(agentValues.filter(agent => agent.selected).map(agent => agent.id)) }} variant={"destructive"} type="submit">Detach Agent</Button>
                                </DialogClose>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                ) : agentValues && agentValues.some(agent => agent.selected) ? (
                    <Dialog>
                        <DialogTrigger>
                            <Button variant={"destructive"}>Detach Agent</Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[425px] h-[16rem]">
                            <DialogHeader>
                                <DialogTitle className="mb-2">Detach Agent</DialogTitle>
                                <DialogDescription>
                                    <p className="text-gray-700 mb-4">Are you sure you want to detach selected agents from {pipelineOverview?.name} Pipeline?</p>
                                    {agentValues.filter(agent => agent.selected).map(
                                        agent => (
                                            <p className="text-gray-600" key={agent.id}>{agent.name}</p>
                                        )
                                    )}
                                </DialogDescription>
                            </DialogHeader>
                            <DialogFooter>
                                <DialogClose className="flex gap-4">
                                    <Button>Cancel</Button>
                                    <Button onClick={() => { handleDetachAgent(agentValues.filter(agent => agent.selected).map(agent => agent.id)) }} variant={"destructive"} type="submit">Detach Agent</Button>
                                </DialogClose>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                ) : (
                    <Dialog>
                        <DialogTrigger asChild>
                            <Button variant="default" className="bg-blue-500">Add Agent</Button>
                        </DialogTrigger>
                        <DialogContent className="sm:max-w-[425px] h-[16rem]">
                            <DialogHeader>
                                <DialogTitle className="mb-2">Add Agent</DialogTitle>
                                <DialogDescription>
                                    <p className="text-gray-700 mb-4">Add an agent from listed agents in the Agents table</p>
                                    <Select onValueChange={(value) => {
                                        const agent = totalAgent.find(agent => agent.name === value);
                                        if (agent) {
                                            setSelectedAgent(agent);
                                        }
                                    }}>
                                        <SelectTrigger className="w-[180px]">
                                            <SelectValue placeholder="Select an agent" />
                                        </SelectTrigger>
                                        <SelectContent>
                                            <SelectGroup>
                                                {connectedAgent && totalAgent && totalAgent
                                                    .filter(agent => !connectedAgent.some(connected => connected.name === agent.name))
                                                    .map(agent => (
                                                        <SelectItem key={agent.id} value={agent.name}>{agent.name}</SelectItem>
                                                    ))}
                                                {!connectedAgent && totalAgent && totalAgent
                                                    .filter(agent => agent.status === "unknown")
                                                    .map(agent => (
                                                        <SelectItem key={agent.id} value={agent.name}>{agent.name}</SelectItem>
                                                    ))}
                                            </SelectGroup>
                                        </SelectContent>
                                    </Select>
                                </DialogDescription>
                            </DialogHeader>
                            <div className="grid gap-4 py-4">
                            </div>
                            <DialogFooter>
                                <DialogClose className="flex gap-4">
                                    <Button>Cancel</Button>
                                    <Button onClick={async () => {
                                        if (selectedAgent) {
                                            await handleAgentApply(selectedAgent);
                                        } else {
                                            console.error("No agent selected to apply.");
                                        }
                                        if (selectedAgent) {
                                            setConnectedAgent([...connectedAgent, {
                                                id: selectedAgent.id,
                                                name: selectedAgent.name,
                                                status: "unknown",
                                                pipeline_name: pipelineOverview?.name || "",
                                                version: selectedAgent.version,
                                                log_rate: selectedAgent.log_rate,
                                                metrics_rate: selectedAgent.metrics_rate,
                                                trace_rate: selectedAgent.trace_rate,
                                                selected: false
                                            }]);
                                        }
                                    }} className="bg-blue-500" type="submit">Apply</Button>
                                </DialogClose>
                            </DialogFooter>
                        </DialogContent>
                    </Dialog>
                )}
            </div>
            {agentValues ? (
                <table className="min-w-full bg-gray-50">
                    <thead>
                        <tr className="border-b border-gray-200">
                            <th className="py-4 px-2 text-left">
                                <input
                                    type="checkbox"
                                    className="h-4 w-4 rounded border-gray-300"
                                    onChange={(e) => {
                                        const isChecked = e.target.checked;
                                        setAgentValues(agentValues.map(device =>
                                            device.pipeline_name === pipelineOverview?.name ? { ...device, selected: isChecked } : device
                                        ));
                                    }}
                                    checked={agentValues
                                        .filter(device => device.pipeline_name === pipelineOverview?.name)
                                        .every(device => device.selected)}
                                />
                            </th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Name</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Status</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Pipeline</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Version</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Log rate</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Metrics Rate</th>
                            <th className="py-4 px-4 text-left font-medium text-gray-600">Trace Rate</th>
                        </tr>
                    </thead>
                    <tbody>
                        {connectedAgent && connectedAgent.map(agent => (
                            <tr key={agent.id} className="border-b border-gray-200">
                                <td className="py-4 px-2">
                                    <input
                                        type="checkbox"
                                        className="h-4 w-4 rounded border-gray-300"
                                        checked={agent.selected}
                                        onChange={() => handleSelectDevice(agent.id)}
                                    />
                                </td>
                                <TableCell className="font-medium text-gray-700">{agent.name}</TableCell>
                                <TableCell className={`${agent.status === "connected" ? "bg-green-100 text-green-700" : agent.status === "disconnected" ? "bg-red-100 text-red-700" : " text-black"}`}>
                                    {agent.status}
                                </TableCell>
                                <TableCell className="text-gray-700">{agent.pipeline_name}</TableCell>
                                <TableCell className="text-gray-700">{agent.version}</TableCell>
                                <TableCell className="text-gray-700">{agent.log_rate}</TableCell>
                                <TableCell className="text-gray-700">{agent.metrics_rate}</TableCell>
                                <TableCell className="text-gray-700">{agent.trace_rate}</TableCell>
                            </tr>
                        ))}
                    </tbody>
                </table>
            ) : (
                <p className="text-gray-600 text-center py-4">No agents available.</p>
            )}
        </div>
    )
}

export default PipelineOverviewTable