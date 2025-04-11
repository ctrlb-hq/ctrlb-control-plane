import { useEffect, useState } from "react"
import { Handle, Position } from "reactflow"
import { Sheet, SheetClose, SheetContent, SheetFooter, SheetTrigger } from "../ui/sheet"
import { Button } from "../ui/button"
import { useNodeValue } from "@/context/useNodeContext"
import usePipelineChangesLog from "@/context/usePipelineChangesLog"

import { JsonForms } from '@jsonforms/react';

import {
  materialCells,
  materialRenderers,
} from '@jsonforms/material-renderers';

import { ThemeProvider, createTheme } from '@mui/material/styles';
import { TransporterService } from "@/services/transporterService"

const theme = createTheme({
  components: {
    MuiFormControl: {
      styleOverrides: {
        root: {
          marginBottom: '0.5rem',
        },
      },
    },
  },
});

const renderers = [
  ...materialRenderers,
];
export const SourceNode = ({ data: Data }: any) => {
  const [isSidebarOpen, setIsSidebarOpen] = useState(false)
  const { setNodeValue } = useNodeValue()
  const { addChange } = usePipelineChangesLog()
  const [form, setForm] = useState<object>({})
  const SourceLabel = Data.supported_signals || ""

  const handleDeleteNode = () => {
    setNodeValue(prev => prev.filter(node => node.id !== Data.component_id));
    setNodeValue(prev => prev.filter(node => node.id !== Data.id));
    const log = { type: 'source', name: Data.name, status: "deleted" }
    const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
    addChange(log)
    const updatedLog = [...existingLog, log];
    localStorage.setItem("changesLog", JSON.stringify(updatedLog));
    const nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
    const updatedNodes = nodes.filter((node: any) => node.component_name !== Data.component_name);
    localStorage.setItem("Nodes", JSON.stringify(updatedNodes));
    setIsSidebarOpen(false);
  }

  const getForm = async () => {
    const res = await TransporterService.getTransporterForm(Data.component_name)
    setForm(res)
  }

  const getSource = JSON.parse(localStorage.getItem("Nodes") || "[]").find((source: any) => source.component_name === Data.component_name);
  const sourceConfig = getSource?.config
  const [data, setData] = useState<object>(Data.config)


  useEffect(() => {
    getForm()
  }, [])

  const handleSubmit = () => {

    const log = { type: 'source', name: Data.name, status: "edited" }
    const existingLog = JSON.parse(localStorage.getItem("changesLog") || "[]");
    addChange(log)
    const updatedLog = [...existingLog, log];
    localStorage.setItem("changesLog", JSON.stringify(updatedLog));

    const nodes = JSON.parse(localStorage.getItem("Nodes") || "[]");
    const updatedNodes = nodes.map((node: any) =>
      node.component_name === Data.component_name ? { ...node, config: data } : node
    );
    localStorage.setItem("Nodes", JSON.stringify(updatedNodes));

    setIsSidebarOpen(false);
  }
  return (
    <Sheet open={isSidebarOpen} onOpenChange={setIsSidebarOpen}>
      <SheetTrigger asChild>
        <div onClick={() => setIsSidebarOpen(true)} className='flex items-center'>
          <div className="flex items-center justify-center rounded-bl-md rounded-tl-md bg-gray-500 h-[6rem]">
            <p className="text-xl m-1 text-white">→|</p>
          </div>
          <div className="bg-gray-200 rounded-tr-md rounded-br-md border-2 border-gray-300 p-4 h-[6rem] shadow-md w-[8rem] relative">
            <div style={{ fontSize: "9px", lineHeight: "0.8rem" }} className="font-medium">{Data.name}</div>
            <div className="flex justify-between gap-2 mr-2 text-xs mt-2">
              {SourceLabel && SourceLabel.map((source: any, index: number) => (
                <p style={{ fontSize: "8px" }} key={index}>
                  {source}
                </p>
              ))}
            </div>
            <Handle type="source" position={Position.Right} className="bg-green-600 w-0 h-0 rounded-full" />
          </div>
          <div className='bg-green-600 h-6 rounded-tr-lg rounded-br-lg w-2' />
        </div>
      </SheetTrigger>
      <SheetContent className="w-[36rem]">
        <div className="flex flex-col gap-4 p-4">
          <div className="flex gap-3 items-center">
            <p className="text-lg bg-gray-500 items-center rounded-lg p-2 px-3 m-1 text-white">→|</p>
            <h2 className="text-xl font-bold">{Data.name}</h2>
          </div>
          <p className="text-gray-500">Generate the defined log type at the rate desired. <span className="text-blue-500 underline">Documentation</span></p>
          <ThemeProvider theme={theme}>
            <div className='mt-3'>
              <div className='text-2xl p-4 font-semibold bg-gray-100'>{form.title}</div>
              <div className='p-3 '>
                <div className='overflow-y-auto h-[29rem]'>
                  <JsonForms
                    data={data}
                    schema={form}
                    renderers={renderers}
                    cells={materialCells}
                    onChange={({ data }) => setData(data)}
                  />
                </div>
              </div>
            </div>
            <SheetFooter>
              <SheetClose>
                <div className="flex gap-3">
                  <Button className="bg-blue-500" onClick={handleSubmit}>Apply</Button>
                  <Button variant={"outline"} onClick={() => setIsSidebarOpen(false)}>Discard Changes</Button>
                  <Button variant={"outline"} onClick={handleDeleteNode}>Delete Node</Button>
                </div>
              </SheetClose>
            </SheetFooter>
          </ThemeProvider>

        </div>
      </SheetContent>
    </Sheet>
  )
}


