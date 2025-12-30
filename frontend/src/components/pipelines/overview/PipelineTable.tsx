import {
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Paper,
} from "@mui/material";
import { useGraphFlow } from "@/context/useGraphFlowContext";
import pipelineServices from "@/services/pipeline";
import { useEffect, useState } from "react";
import ViewPipelineDetails from "./ViewPipelineDetails";

interface Pipeline {
  id: string;
  name: string;
  agents: number;
  incoming_bytes: number;
  outgoing_bytes: number;
  updatedAt: number;
}

const formatTimestamp = (timestamp: number) =>
  new Date(timestamp * 1000)
    .toLocaleString("en-GB", {
      day: "2-digit",
      month: "2-digit",
      year: "numeric",
      hour: "2-digit",
      minute: "2-digit",
      second: "2-digit",
      hour12: false,
    })
    .replace(",", "");

const PipelineTable = () => {
  const [pipelines, setPipelines] = useState<Pipeline[]>([]);
  const [selectedPipelineId, setSelectedPipelineId] = useState<string | null>(null);
  const { resetGraph } = useGraphFlow();

  const handleGetPipelines = async () => {
    const res = await pipelineServices.getAllPipelines();
    setPipelines(res);
  };

  useEffect(() => {
    handleGetPipelines();
  }, []);

  const handleCloseDetails = () => {
    setSelectedPipelineId(null);
    resetGraph();
    handleGetPipelines();
  };

  return (
    <>
      <TableContainer component={Paper} variant="outlined">
        <Table>
          <TableHead>
            <TableRow sx={{ backgroundColor: "#f5f5f5" }}>
              <TableCell>Name</TableCell>
              <TableCell>Incoming bytes</TableCell>
              <TableCell>Outgoing bytes</TableCell>
              <TableCell>Updated at</TableCell>
            </TableRow>
          </TableHead>

          <TableBody>
            {pipelines.map(pipeline => (
              <TableRow
                key={pipeline.id}
                hover
                sx={{ cursor: "pointer" }}
                onClick={() => setSelectedPipelineId(pipeline.id)}
              >
                <TableCell>{pipeline.name}</TableCell>
                <TableCell>{pipeline.incoming_bytes}</TableCell>
                <TableCell>{pipeline.outgoing_bytes}</TableCell>
                <TableCell>{formatTimestamp(pipeline.updatedAt)}</TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {selectedPipelineId && (
        <ViewPipelineDetails
          pipelineId={selectedPipelineId}
          open={Boolean(selectedPipelineId)}
          onClose={handleCloseDetails}
        />
      )}
    </>
  );
};

export default PipelineTable;
