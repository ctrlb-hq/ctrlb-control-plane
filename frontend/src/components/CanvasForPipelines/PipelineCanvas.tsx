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

import 'reactflow/dist/style.css';
import { SourceNode } from './SourceNode';
import { ProcessorNode } from './ProcessorNode';
import { DestinationNode } from './DestinationNode';
import SourceDropdownOptions from '../Pipelines/DropdownOptions/SourceDropdownOptions';
import ProcessorDropdownOptions from '../Pipelines/DropdownOptions/ProcessorDropdownOptions';
import DestinationDropdownOptions from '../Pipelines/DropdownOptions/DestinationDropdownOptions';
import { useNodeValue } from '@/context/useNodeContext';


// Node types mapping
const nodeTypes = {
  source: SourceNode,
  processor: ProcessorNode,
  destination: DestinationNode
};

const PipelineCanvas = () => {

  const { nodeValue, setNodeValue, onNodesChange } = useNodeValue();

  const validatedNodeValue = nodeValue.map((node, index) => ({
    ...node,
    position: node.position || { x: 100, y: 100 + index * 100 },
  }));

  const [edges, setEdges, onEdgesChange] = useEdgesState(JSON.parse(localStorage.getItem("PipelineEdges") || "[]"));
  const [reactFlowInstance, setReactFlowInstance] = useState<ReactFlowInstance | null>(null);

  const onConnect = useCallback(
    (params: Edge | Connection) => {
      setEdges((eds) => {
        if (!params.source || !params.target) {
          console.error('Invalid connection: source or target is null');
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
      if (!type) return;

      if (!reactFlowInstance) return;
      const position = reactFlowInstance.project({ x: event.clientX, y: event.clientY });
      let nodeData;
      const id = `node_${Date.now()}`;

      const newNode = { id, type, position, data: nodeData };
      setNodeValue((nds) => nds.concat(newNode));
    },
    [reactFlowInstance, nodeValue, setNodeValue]
  );

  return (
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
  );
};

export default PipelineCanvas;
