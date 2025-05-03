import React, { useState, useCallback, } from 'react';
import ReactFlow, {
  MiniMap,
  Controls,
  Background,
  addEdge,
  useEdgesState,
  Edge,
  Connection,
  ReactFlowInstance,
} from 'reactflow';

import {
  Sheet,
  SheetClose,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet"

import 'reactflow/dist/style.css';
import { SourceNode } from './SourceNode';
import { ProcessorNode } from './ProcessorNode';
import { DestinationNode } from './DestinationNode';
import SourceDropdownOptions from '../Pipelines/DropdownOptions/SourceDropdownOptions';
import ProcessorDropdownOptions from '../Pipelines/DropdownOptions/ProcessorDropdownOptions';
import DestinationDropdownOptions from '../Pipelines/DropdownOptions/DestinationDropdownOptions';
import { useNodeValue } from '@/context/useNodeContext';
import { Button } from '../ui/button';
import { Edit } from 'lucide-react';
import { Label } from "@/components/ui/label"
import { Switch } from "@/components/ui/switch"
import usePipelineChangesLog from '@/context/usePipelineChangesLog';
import pipelineServices from '@/services/pipelineServices';
import { toast } from '@/hooks/use-toast';


// Node types mapping
const nodeTypes = {
  source: SourceNode,
  processor: ProcessorNode,
  destination: DestinationNode
};

const PipelineCanvas = () => {

  const { nodeValue, setNodeValue, onNodesChange } = useNodeValue();

  const validatedNodeValue = nodeValue.map((node, index) => {
    const nodeType = node.type;
    let x, y;
    if (!node.position || node.position.x == -1 || node.position.y == -1) {
      if (nodeType === 'source') {
        x = 50; // Fixed left position
        y = 50 + (index * 100);
      } else if (nodeType === 'destination') {
        x = 400; // Fixed right position
        y = 50 + (index * 100);
      } else { // processor
        x = 225; // Center position
        y = 50 + (index * 100);
      }
      return {
        ...node,
        position: { x, y }
      };
    }
    return node;
  });

  const [edges, setEdges, onEdgesChange] = useEdgesState(JSON.parse(localStorage.getItem("PipelineEdges") || "[]"));
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);
  const [check, setCheck] = useState(true)
  const { changesLog } = usePipelineChangesLog()

  const pipelineName = localStorage.getItem('pipelinename');
  const createdBy = localStorage.getItem('userEmail');
  const agentIds = JSON.parse(localStorage.getItem('selectedAgentIds') || '');
  const PipelineNodes = JSON.parse(localStorage.getItem('Nodes') || '[]');
  const PipelineEdges = JSON.parse(localStorage.getItem('PipelineEdges') || '[]') || [];

  const onConnect = useCallback(
    (params: Edge | Connection) => {
      setEdges((eds) => {
        if (!params.source || !params.target) {
          console.error('Invalid connection: source or target is null');
          return eds;
        }
        //check the node corresponding to the source and target
        const sourceNode = nodeValue.find((node) => node.id === params.source);
        const targetNode = nodeValue.find((node) => node.id === params.target);
        if (!sourceNode || !targetNode) {
          console.error('Invalid connection: source or target node not found');
          return eds;
        }

        //check if the source and target are of the same type
        if (sourceNode.type === targetNode.type) {
          console.error('Invalid connection: source and target are of the same type');
          return eds;
        }

        //check if the source and target compatibility, ie the supported signals in source must be a subset of the supported signals in target
        if (!targetNode.data.supported_signals.some((signal: string) => sourceNode.data.supported_signals.includes(signal))) {
          console.error('Invalid connection: source and target are not compatible');
          toast({
            title: "Error",
            description: "Source and target are not compatible",
            variant: "destructive",
          });
          return eds;
        }

        const updatedEdges = addEdge(
          {
            ...params,
            source: params.source,
            target: params.target,
            animated: true,
            data: {
              sourceComponentId: parseInt(params.source, 10),
              targetComponentId: parseInt(params.target, 10),
            },
          },
          eds
        );
        localStorage.setItem('PipelineEdges', JSON.stringify(updatedEdges));
        return updatedEdges;
      });
    },
    [setEdges]
  );


  const onDragOver = useCallback((event: React.DragEvent) => {
    event.preventDefault();
    event.dataTransfer.dropEffect = 'move';
  }, []);

  const onDrop = useCallback(
    (event: React.DragEvent) => {
      event.preventDefault();
      const type = event.dataTransfer.getData('application/nodeType');
      if (!type || !reactFlowInstance) return;
      const position = reactFlowInstance.project({ x: event.clientX, y: event.clientY });
      let nodeData;
      const id = `node_${Date.now()}`;

      const newNode = { id, type, position, data: nodeData };
      setNodeValue((nds) => nds.concat(newNode));
    },
    [reactFlowInstance, nodeValue, setNodeValue]
  );

  const addPipeline = async () => {
    console.log("PipelineNodes", PipelineNodes)
    const pipelinePayload = {
      "name": pipelineName,
      "created_by": createdBy,
      "agent_ids": [parseInt(agentIds)],
      "pipeline_graph": {
        "nodes": PipelineNodes.map((node: { id: string; data: { name: any; component_name: any; config: any; supported_signals: any; }; type: string; }) => ({
          component_id: parseInt(node.id),
          name: node.data.name,
          component_role: node.type === 'source' ? 'exporter' : node.type === 'destination' ? 'receiver' : 'processor',
          component_name: node.data.component_name,
          supported_signals: node.data.supported_signals,
          config: node.data.config
        })),
        "edges": JSON.parse(localStorage.getItem('PipelineEdges') || '[]').map((edge: { source: any; target: any; }) => ({
          source: edge.source,
          target: edge.target
        }))
      }
    }
    console.log("payload", pipelinePayload)
    const res = await pipelineServices.addPipeline(pipelinePayload)
    console.log(res)
  }

  const handleDeployChanges = () => {
    addPipeline()
    localStorage.removeItem('Sources');
    localStorage.removeItem('Destination');
    localStorage.removeItem('pipelinename');
    localStorage.removeItem("selectedAgentIds")
    localStorage.removeItem("Nodes")
    localStorage.removeItem("changesLog")
    localStorage.removeItem("PipelineEdges")
    setTimeout(() => {
      toast({
        title: "Success",
        description: "Successfully deployed the pipeline",
        duration: 3000,

      });
      window.location.reload()
    }, 2000);
  }



  return (
    <>
      <SheetContent>
        <SheetHeader>
          <div className="flex justify-between items-center p-2 border-b">
            <SheetTitle>
              <div className="flex items-center space-x-2">
                <div className="text-xl font-medium">{pipelineName}</div>
              </div>
            </SheetTitle>

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
                              {/* <p className="text-lg capitalize">{change.component_role}</p> */}
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
                      <Button onClick={handleDeployChanges} className="bg-blue-500">Deploy Changes</Button>
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
        </SheetHeader>
        <div className="w-full flex flex-col gap-2 h-screen p-4">
          <div className="h-4/5 border-2 border-gray-200 rounded-lg">
            <ReactFlow
              nodes={validatedNodeValue}
              edges={edges}
              onNodesChange={onNodesChange}
              onEdgesChange={onEdgesChange}
              onConnect={onConnect}
              onInit={setReactFlowInstance}
              onDrop={onDrop}
              onDragOver={onDragOver}
              nodeTypes={nodeTypes}
              fitView
            >
              <MiniMap />
              <Controls />
              <Background color="#aaa" gap={16} />
            </ReactFlow>
          </div>
          <div className=' p-2 pb-4'>
            <div className="flex justify-center gap-6 items-center">
              <div className="flex gap-6 bg-gray-100 p-4 rounded-lg">
                <SourceDropdownOptions />
                <ProcessorDropdownOptions />
                <DestinationDropdownOptions />
              </div>
            </div>
          </div>

        </div>
      </SheetContent>
    </>

  );
};

export default PipelineCanvas;
