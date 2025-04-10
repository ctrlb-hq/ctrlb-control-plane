import ProgressFlow from './ProgressFlow'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardFooter } from '@/components/ui/card'
import { usePipelineStatus } from '@/context/usePipelineStatus';
import { useEffect, useState } from 'react';
import { ChevronUp, Code2, Edit, Loader2 } from 'lucide-react';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import { useToast } from '@/hooks/use-toast';
import Tabs from './Tabs';
import PipelineCanvas from '@/components/CanvasForPipelines/PipelineCanvas';
import { useAgentValues } from '@/context/useAgentsValues';
import { AgentValuesTable } from '@/types/agentValues.type';
import { usePipelineTab } from '@/context/useAddNewPipelineActiveTab';
import CreateNewAgent from '@/components/Agents/CreateNewAgent';
import pipelineServices from '@/services/pipelineServices';
import usePipelineChangesLog from '@/context/usePipelineChangesLog';
import { NodeValueProvider } from '@/context/useNodeContext';
import { pipe } from 'framer-motion';


const AddAgent = () => {
  const pipelineStatus = usePipelineStatus();
  if (!pipelineStatus) {
    return null;
  }
  let { currentStep, setCurrentStep } = pipelineStatus;
  const [selectedRows, setSelectedRows] = useState<string[]>([]);
  const [sortDirection, setSortDirection] = useState('asc');
  const [isDialogOpen, setIsDialogOpen] = useState(false);
  const [selectedAgents, setSelectedAgents] = useState<AgentValuesTable[]>([]);
  const { changesLog } = usePipelineChangesLog()
  const [rollOut, setRollOut] = useState(false)
  const { toast } = useToast()
  const { agentValues } = useAgentValues()
  const [check, setCheck] = useState(true)
  const { currentTab } = usePipelineTab()
  const [filteredAgents, setFilteredAgents] = useState<AgentValuesTable[]>([]);

  useEffect

  const pipelineName = localStorage.getItem('pipelinename');
  const createdBy = localStorage.getItem('userEmail');
  const agentIds = JSON.parse(localStorage.getItem('selectedAgentIds') || '[]');
  const PipelineNodes = JSON.parse(localStorage.getItem('Nodes') || '[]');
  const PipelineEdges = JSON.parse(localStorage.getItem('PipelineEdges') || '[]') || [];



  const addPipeline = async () => {
    const pipelinePayload = {
      "name": pipelineName,
      "created_by": createdBy,
      "agent_ids": agentIds,
      "pipeline_graph": {
        "nodes": PipelineNodes.map(node => ({
          ...node.data,
          component_id: Number(node.data.component_id) // Ensure component_id is a number
        })),
        "edges": PipelineEdges
      }
    }
    console.log("edges: ", pipelinePayload.pipeline_graph.edges)
    console.log("payload", pipelinePayload)
    const res = await pipelineServices.addPipeline(pipelinePayload)
    console.log(res)
  }

  useEffect(() => {
    const storedAgentIds = JSON.parse(localStorage.getItem("selectedAgentIds") || "[]");
    if (storedAgentIds.length > 0) {
      setSelectedRows(storedAgentIds); // Pre-select rows based on stored IDs
      const preSelectedAgents = agentValues.filter(agent => storedAgentIds.includes(agent.id));
      setSelectedAgents(preSelectedAgents); // Pre-select agent objects
    }
  }, [agentValues]);

  const toggleSelectAll = () => {
    if (selectedRows.length === agentValues.length) {
      setSelectedRows([]);
      setSelectedAgents([]);
    } else {
      setSelectedRows(agentValues.map(agent => agent.id));
      setSelectedAgents(agentValues);
    }
  };

  const toggleSelectRow = (id: string) => {
    if (selectedRows.includes(id)) {
      setSelectedRows(selectedRows.filter(rowId => rowId !== id));
      setSelectedAgents(selectedAgents.filter(agent => agent.id !== id));
    } else {
      setSelectedRows([...selectedRows, id]);
      const selectedAgent = agentValues.find(agent => agent.id === id);
      if (selectedAgent) {
        setSelectedAgents([...selectedAgents, selectedAgent]);
      }
    }
  };

  const handleSort = () => {
    setSortDirection(sortDirection === 'asc' ? 'desc' : 'asc');
  };


  const handleApply = () => {
    const selectedAgentsData = agentValues.filter(agent => selectedRows.includes(agent.id));
    setSelectedAgents(selectedAgentsData);
    setIsDialogOpen(false);
    const selectedAgentIds = selectedAgentsData.map(agent => agent.id);
    localStorage.setItem('selectedAgentIds', JSON.stringify(selectedAgentIds));
  };

  const handleRollout = () => {
    setRollOut(true);
    setTimeout(() => {
      setRollOut(false);
      toast({
        title: "Success",
        description: "Rolled out successfully",
        duration: 3000,
      });
    }, 2000);
  }

  const handleEditAgent = (agent: AgentValuesTable) => {
    setSelectedAgents([agent])
    setIsDialogOpen(true)
  }

  const handleGetAgent = async () => {
    setFilteredAgents(agentValues);
  };

  useEffect(() => {
    if (localStorage.getItem('authToken'))
      handleGetAgent();
  }, []);

  const handleSearch = (e: React.ChangeEvent<HTMLInputElement>) => {
    const searchValue = e.target.value.toLowerCase();
    const filtered = agentValues.filter(
      (agent) =>
        agent.name.toLowerCase().includes(searchValue) ||
        agent.status.toLowerCase().includes(searchValue) ||
        agent.version.toLowerCase().includes(searchValue)
    );
    setFilteredAgents(filtered);
  };

  const handleDeployChanges = () => {
    addPipeline()
    localStorage.removeItem('Sources');
    localStorage.removeItem('Destination');
    localStorage.removeItem('pipelinename');
    localStorage.removeItem("selectedAgentIds")
    localStorage.removeItem("Nodes")
    localStorage.removeItem("changesLog")
    setTimeout(() => {
      toast({
        title: "Success",
        description: "Successfully deployed the pipeline",
        duration: 3000,

      });
      localStorage.removeItem("PipelineEdges")

      // window.location.reload()
    }, 2000);
  }


  return (
    <div className='flex flex-col gap-5'>
      <Tabs />
      {currentTab == "pipelines" ? <div className="mx-auto flex gap-5 w-full">
        <ProgressFlow />
        <Card className="w-full h-[40rem] bg-white shadow-sm">
          <CardHeader>
            <div className='flex flex-col gap-2'>
              <p className='text-xl'>Add Agent</p>
              <p className='text-md text-gray-600 mb-4'>Agents Collect Data from the sources in the pipeline and route them to desired destination.</p>
            </div>
          </CardHeader>
          <CardContent className='h-[29rem] w-full'>
            <div className='flex flex-col '>{selectedAgents.length > 0 ? (
              <ul>
                {selectedAgents.map((agent, index) => (
                  <div key={index} className='flex border justify-between rounded-md mb-2 p-3 border-gray-00 gap-4'>
                    <div className='flex justify-start gap-2 items-center'>
                      <Code2 className='text-gray-500' />
                      <p className=''>{agent.name}</p>
                    </div>

                    <div className='flex gap-2 justify-end items-center'>
                      <Button onClick={() => handleEditAgent(agent)} className='bg-blue-500'>
                        Edit
                      </Button>
                      <Button
                        variant={"destructive"}
                        onClick={() => {
                          const updatedAgents = selectedAgents.filter((_, i) => i !== index);
                          setSelectedAgents(updatedAgents);
                          setSelectedRows(updatedAgents.map(agent => agent.id));
                        }}
                      >
                        Delete
                      </Button>
                    </div>
                  </div>
                ))}
              </ul>
            ) : (
              <p className="text-gray-500">No agents selected.</p>
            )}</div>
            {selectedAgents.length > 0 && <Button disabled={rollOut} onClick={handleRollout} className='bg-blue-500 mb-5 w-full mt-2'>RollOut</Button>}
            {rollOut && <div className='flex border border-blue-400 p-5 mb-2 rounded-md justify-center items-center gap-2 text-blue-700'>
              <Loader2 className='animate-spin' />
              <p>Rollout in progress, we will notify once it's completed mean while you can navigate to pipeline map</p>
            </div>}
            <Dialog open={isDialogOpen} onOpenChange={setIsDialogOpen}>
              <DialogTrigger asChild>
                <div className='mt-2'>
                  <Button className='w-full' variant="outline">Add Agent</Button>
                </div>
              </DialogTrigger>
              <DialogContent className="w-full h-screen">
                <DialogHeader>
                  <DialogTitle className='text-xl'>Apply Configuration</DialogTitle>
                  <DialogDescription>
                    <div className="w-full mt-5 border rounded-md shadow-sm h-[38rem] bg-white">
                      <div className="p-4 flex gap-2 border-b">
                        <Input
                          placeholder="Search by name"
                          className="flex-1"
                          onChange={handleSearch}
                        />
                      </div>
                      <div className='flex flex-col'>
                        <div className="w-full h-[29rem] overflow-auto">
                          <table className="w-full  text-sm">
                            <thead>
                              <tr className="border-b bg-gray-50">
                                {agentValues && (
                                  <th className="px-4 py-3 text-left w-12">
                                    <Checkbox
                                      checked={selectedRows.length === agentValues.length}
                                      onCheckedChange={toggleSelectAll}
                                    />
                                  </th>
                                )}
                                <th
                                  className="px-4 py-3 text-left font-medium cursor-pointer"
                                  onClick={handleSort}
                                >
                                  Name
                                  <ChevronUp className={`w-4 h-4 inline ml-1 ${sortDirection === 'asc' ? 'opacity-100' : 'opacity-30'}`} />
                                </th>
                                <th className="px-4 py-3 text-left font-medium">Status</th>
                                <th className="px-4 py-3 text-left font-medium">Version</th>
                                <th className="px-4 py-3 text-left font-medium">Log Rate</th>
                                <th className="px-4 py-3 text-left font-medium">Trace Rate</th>
                                <th className="px-4 py-3 text-left font-medium">Metrics Rate</th>
                              </tr>
                            </thead>
                            <tbody>
                              {filteredAgents && filteredAgents.filter(agent => agent.pipeline_name == "").map((agent) => (
                                <tr key={agent.id} className="border-b hover:bg-gray-50">
                                  <td className="px-4 py-3">
                                    <Checkbox
                                      checked={selectedRows.includes(agent.id)}
                                      onCheckedChange={() => toggleSelectRow(agent.id)}
                                    />
                                  </td>
                                  <td className="px-4 py-3 font-medium text-blue-500">
                                    {agent.name}
                                  </td>
                                  <td className='px-4 py-3'>
                                    <p className='bg-green-700 rounded-full flex justify-center items-center p-1 text-white'>{agent.status}</p>
                                  </td>
                                  <td className="px-4 py-3">{agent.version}</td>
                                  <td className="px-4 py-3">{agent.log_rate}</td>
                                  <td className="px-4 py-3">{agent.trace_rate}</td>
                                  <td className="px-4 py-3">{agent.metrics_rate}</td>
                                </tr>
                              ))}
                            </tbody>
                          </table>
                        </div>
                      </div>
                    </div>
                  </DialogDescription>
                </DialogHeader>
                <DialogFooter>
                  <div className="p-4 border-t flex justify-end gap-2 pb-5">
                    <Button variant="outline" onClick={() => setIsDialogOpen(false)}>Cancel</Button>
                    <Button className="bg-blue-500 hover:bg-blue-600 text-white" onClick={handleApply}>Apply</Button>
                  </div>
                </DialogFooter>
              </DialogContent>
            </Dialog>
          </CardContent>
          <CardFooter className="flex justify-end items-end">
            <div className=" flex items-end justify-end gap-4">
              <Button
                className="bg-gray-700 px-6 disabled:opacity-50"
                disabled={currentStep === 0}
                onClick={() => setCurrentStep(--currentStep)}
              >
                Back
              </Button>
              <Sheet>
                <SheetTrigger>
                  <Button
                    className="bg-blue-500 hover:bg-blue-700 px-6 disabled:opacity-50"
                    onClick={() => setCurrentStep(++currentStep)}
                  >
                    {selectedAgents.length > 0 ? "Save & View Pipeline" : "Skip & Save Pipeline"}
                  </Button>
                </SheetTrigger>
                <SheetContent>
                  <SheetHeader>
                    <SheetTitle>
                      <div className="flex justify-between items-center p-4 border-b">
                        <div className="flex items-center space-x-2">
                          <div className="text-xl font-medium">{pipelineName}</div>
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
                                          <p className="text-lg capitalize">{change.component_role}</p>
                                          <p className="text-lg text-gray-800 capitalize">{change.name}</p>
                                        </div>
                                        <div className="flex justify-end gap-3 items-center">
                                          <p className={`${change.status == 'edited' ? "text-gray-500" : change.status == 'deleted' ? "text-red-500" : "text-green-600"} text-lg`}>[{change.status ? change.status : "Added"}]</p>
                                          <Edit size={20} />
                                        </div>
                                      </div>
                                    ))
                                  }
                                </div>
                              </SheetDescription>
                              <SheetClose className="flex justify-end mt-4 w-full">
                                <div>
                                  <Button onClick={handleDeployChanges} className="bg-blue-500">Deploy Changes1</Button>
                                </div>
                              </SheetClose>
                            </SheetContent>
                          </Sheet>
                          <div className="mx-4 flex items-center space-x-2">
                            <Switch id="edit-mode" checked={check} onCheckedChange={setCheck} />
                            <Label htmlFor="edit-mode">Edit Mode</Label>
                          </div>
                        </div>
                      </div>
                    </SheetTitle>
                    <SheetDescription>
                      <NodeValueProvider>
                        <PipelineCanvas />

                      </NodeValueProvider>
                    </SheetDescription>
                  </SheetHeader>
                </SheetContent>
              </Sheet>
            </div>
          </CardFooter>
        </Card>

      </div> : <CreateNewAgent />}
    </div>
  )
}

export default AddAgent;