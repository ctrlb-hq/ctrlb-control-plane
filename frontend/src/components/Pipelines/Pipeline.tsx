import {
  Table,
  TableBody,
  TableCaption,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import LandingView from "./LandingView";
import {
  Sheet,
  SheetContent,
  SheetTrigger,
} from "@/components/ui/sheet"
import PipelineOverview from "./PipelineOverview";
import pipelineServices from "@/services/pipelineServices";
import { useEffect, useState } from "react";
import { usePipelineOverview } from "@/context/usePipelineDetailContext";


interface pipeline {
  id: string,
  name: string,
  agents: number,
  incoming_bytes: number,
  outgoing_bytes: number,
  updatedAt: number, 
}

const Pipeline = () => {
  const [pipelines, setPipelines] = useState<pipeline[]>([])
  const { setPipelineOverview } = usePipelineOverview()
  const [pipelineId, setPipelineId] = useState<string>("")
  const handleGetPipelines = async () => {
    const res = await pipelineServices.getAllPipelines()
    setPipelines(res)
  }

  const handleGetPipeline = async () => {
    const res = await pipelineServices.getPipelineById(pipelineId)
    setPipelineOverview(res)
  }

  useEffect(() => {
    handleGetPipelines()
    handleGetPipeline()
  }, [])

  const formatTimestamp = (timestamp: number) => {
    const date = new Date(timestamp * 1000) // Convert seconds to milliseconds
    const hours = date.getHours().toString().padStart(2, '0')
    const minutes = date.getMinutes().toString().padStart(2, '0')
    return `${hours}:${minutes}`
  }

  return (
    <>
      {pipelines && (
        <Table className="border border-gray-200">
          <TableCaption>A list of your recent pipelines.</TableCaption>
          <TableHeader className="bg-gray-100">
            <TableRow>
              <TableHead className="w-[100px]">Name</TableHead>
              <TableHead className="w-[100px]">Agents</TableHead>
              <TableHead className="w-[100px]">Incoming bytes</TableHead>
              <TableHead className="w-[100px]">Outgoing bytes</TableHead>
              <TableHead className="w-[100px]">Updated at</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {Array.isArray(pipelines)&& pipelines.map((pipeline) => (
              <Sheet key={pipeline.id}>
                <SheetTrigger asChild>
                  <TableRow className="cursor-pointer" key={pipeline.id} onClick={() => setPipelineId(pipeline.id)}>
                    <TableCell className="font-medium text-gray-700">{pipeline.name}</TableCell>
                    <TableCell className="text-gray-700">{pipeline.agents}</TableCell>
                    <TableCell className="text-gray-700">{pipeline.incoming_bytes}</TableCell>
                    <TableCell className="text-gray-700">{pipeline.outgoing_bytes}</TableCell>
                    <TableCell className="text-gray-700">{formatTimestamp(pipeline.updatedAt)}</TableCell>
                  </TableRow>
                </SheetTrigger>
                <SheetContent>
                  <PipelineOverview pipelineId={pipeline.id} />
                </SheetContent>
              </Sheet>
            ))}
          </TableBody>
        </Table>
      )}
      {!pipelines && <LandingView />}
    </>
  )
}

export default Pipeline