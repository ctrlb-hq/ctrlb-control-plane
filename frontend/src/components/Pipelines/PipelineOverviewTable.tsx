import { TableCell } from "@/components/ui/table";
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
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectGroup,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useEffect, useState } from "react";
import pipelineServices from "@/services/pipelineServices";
import { Agents } from "@/types/agent.types";

const PipelineOverviewTable = ({ pipelineId }: { pipelineId: string }) => {
  const { pipelineOverview } = usePipelineOverview();

  // State for attached agents and unattached agents
  const [attachedAgents, setAttachedAgents] = useState<Agents[]>([]);
  const [unattachedAgents, setUnattachedAgents] = useState<Agents[]>([]);
  const [selectedAgent, setSelectedAgent] = useState<Agents | null>(null);

  // Fetch both attached and unattached agents from their respective APIs.
  const fetchAgents = async () => {
    const authToken = localStorage.getItem("authToken");
    if (!authToken) {
      console.error("Unauthorized: No authToken found. Skipping agent fetch.");
      return;
    }
    try {
      // Fetch attached agents from the pipeline.
      const attached = await pipelineServices.getAllAgentsAttachedToPipeline(
        pipelineId
      );
      // Ensure each attached agent has a "selected" property.
      const attachedWithSelected = Array.isArray(attached)
        ? attached.map((agent: Agents) => ({
            ...agent,
            selected: agent.selected ?? false,
          }))
        : [];
      setAttachedAgents(attachedWithSelected);

      // Fetch unattached agents via the provided API.
      const unattached = await pipelineServices.getAllUnattachedAgents();
      setUnattachedAgents(Array.isArray(unattached) ? unattached : []);
    } catch (error) {
      console.error("Error fetching agents:", error);
    }
  };

  useEffect(() => {
    fetchAgents();
    // Refresh when pipelineId changes.
  }, [pipelineId]);

  // Toggle the "selected" state for an agent in the attached list.
  const handleSelectDevice = (id: string) => {
    setAttachedAgents(
      attachedAgents.map((agent) =>
        agent.id === id ? { ...agent, selected: !agent.selected } : agent
      )
    );
  };

  // Attach an agent using the API then refresh the agent lists.
  const handleAgentApply = async (agent: Agents) => {
    try {
      await pipelineServices.attachAgentToPipeline(pipelineId, agent.id);
      await fetchAgents();
    } catch (error) {
      console.error("Failed to attach agent:", error);
    }
  };

  // Detach one or more agents then refresh the agent lists.
  const handleDetachAgent = async (ids: string[]) => {
    try {
      for (const id of ids) {
        await pipelineServices.detachAgentFromPipeline(pipelineId, id);
      }
      await fetchAgents();
    } catch (error) {
      console.error("Failed to detach agents:", error);
    }
  };

  // Helper to return CSS classes based on an agent’s status.
  const getStatusClass = (status: string) => {
    if (status === "connected") return "bg-green-100 text-green-700";
    if (status === "disconnected") return "bg-red-100 text-red-700";
    return "text-black";
  };

  // Identify which attached agents are selected.
  const selectedAgentIds = attachedAgents
    .filter((agent) => agent.selected)
    .map((agent) => agent.id);
  const allSelected =
    attachedAgents.length > 0 &&
    attachedAgents.every((agent) => agent.selected);

  return (
    <div className="p-4 rounded-lg shadow">
      <div className="flex mb-5 justify-between">
        <h1 className="text-xl flex justify-center items-center text-gray-600">
          Agents
          <RefreshCcwIcon
            className="w-5 mx-4 text-blue-500 cursor-pointer"
            onClick={fetchAgents}
          />
        </h1>

        {selectedAgentIds.length > 0 ? (
          // Detach dialog appears if one or more agents are selected.
          <Dialog>
            <DialogTrigger>
              <Button variant="destructive">
                {allSelected ? "Detach Agents" : "Detach Agent"}
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px] h-[16rem]">
              <DialogHeader>
                <DialogTitle className="mb-2">
                  {allSelected ? "Detach All Agents" : "Detach Agent"}
                </DialogTitle>
                <DialogDescription>
                  <p className="text-gray-700 mb-4">
                    Are you sure you want to detach the selected agents from{" "}
                    {pipelineOverview?.name} Pipeline?
                  </p>
                  {attachedAgents
                    .filter((agent) => agent.selected)
                    .map((agent) => (
                      <p className="text-gray-600" key={agent.id}>
                        {agent.name}
                      </p>
                    ))}
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose className="flex gap-4">
                  <Button>Cancel</Button>
                  <Button
                    onClick={() => handleDetachAgent(selectedAgentIds)}
                    variant="destructive"
                    type="submit"
                  >
                    Detach Agent
                  </Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        ) : (
          // Add Agent dialog appears if no agent is selected.
          <Dialog>
            <DialogTrigger asChild>
              <Button variant="default" className="bg-blue-500">
                Add Agent
              </Button>
            </DialogTrigger>
            <DialogContent className="sm:max-w-[425px] h-[16rem]">
              <DialogHeader>
                <DialogTitle className="mb-2">Add Agent</DialogTitle>
                <DialogDescription>
                  <p className="text-gray-700 mb-4">
                    Add an agent from the list of unattached agents.
                  </p>
                  <Select
                    onValueChange={(value) => {
                      const agent = unattachedAgents.find(
                        (agent) => agent.name === value
                      );
                      if (agent) setSelectedAgent(agent);
                    }}
                  >
                    <SelectTrigger className="w-[180px]">
                      <SelectValue placeholder="Select an agent" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectGroup>
                        {unattachedAgents && unattachedAgents.length > 0 ? (
                          unattachedAgents.map((agent) => (
                            <SelectItem key={agent.id} value={agent.name}>
                              {agent.name}
                            </SelectItem>
                          ))
                        ) : (
                          // Show disabled option if no unattached agents
                          <SelectItem disabled value="no-agents-available">
                            No agents available
                          </SelectItem>
                        )}
                      </SelectGroup>
                    </SelectContent>
                  </Select>
                </DialogDescription>
              </DialogHeader>
              <DialogFooter>
                <DialogClose className="flex gap-4">
                  <Button>Cancel</Button>
                  <Button
                    className="bg-blue-500"
                    type="submit"
                    onClick={async () => {
                      if (selectedAgent) await handleAgentApply(selectedAgent);
                    }}
                    disabled={!selectedAgent}
                  >
                    Apply
                  </Button>
                </DialogClose>
              </DialogFooter>
            </DialogContent>
          </Dialog>
        )}
      </div>

      {attachedAgents.length > 0 ? (
        <table className="min-w-full bg-gray-50">
          <thead>
            <tr className="border-b border-gray-200">
              <th className="py-4 px-2 text-left">
                <input
                  type="checkbox"
                  className="h-4 w-4 rounded border-gray-300"
                  onChange={(e) => {
                    const isChecked = e.target.checked;
                    setAttachedAgents(
                      attachedAgents.map((agent) =>
                        agent.pipeline_name === pipelineOverview?.name
                          ? { ...agent, selected: isChecked }
                          : agent
                      )
                    );
                  }}
                  checked={
                    attachedAgents.length > 0 &&
                    attachedAgents.every((agent) => agent.selected)
                  }
                />
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Name
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Status
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Pipeline
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Version
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Log rate
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Metrics Rate
              </th>
              <th className="py-4 px-4 text-left font-medium text-gray-600">
                Trace Rate
              </th>
            </tr>
          </thead>
          <tbody>
            {attachedAgents.map((agent) => (
              <tr key={agent.id} className="border-b border-gray-200">
                <td className="py-4 px-2">
                  <input
                    type="checkbox"
                    className="h-4 w-4 rounded border-gray-300"
                    checked={agent.selected}
                    onChange={() => handleSelectDevice(agent.id)}
                  />
                </td>
                <TableCell className="font-medium text-gray-700">
                  {agent.name}
                </TableCell>
                <TableCell className={getStatusClass(agent.status)}>
                  {agent.status}
                </TableCell>
                <TableCell className="text-gray-700">
                  {agent.pipeline_name}
                </TableCell>
                <TableCell className="text-gray-700">{agent.version}</TableCell>
                <TableCell className="text-gray-700">
                  {agent.log_rate}
                </TableCell>
                <TableCell className="text-gray-700">
                  {agent.metrics_rate}
                </TableCell>
                <TableCell className="text-gray-700">
                  {agent.trace_rate}
                </TableCell>
              </tr>
            ))}
          </tbody>
        </table>
      ) : (
        <p className="text-gray-600 text-center py-4">No agents attached.</p>
      )}
    </div>
  );
};

export default PipelineOverviewTable;
