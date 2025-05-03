import Editor from "@monaco-editor/react";

import { Card, CardBody, Typography, Button } from "@material-tailwind/react";

function AgentDetails() {
  const details = [
    { label: "Name", value: "gallium" },
    { label: "Status", value: "Connected" },
    { label: "Type", value: "BindPlane Agent" },
    { label: "Version", value: "v1.60.0" },
    { label: "Host Name", value: "gallium" },
    { label: "Platform", value: "linux amd64" },
    { label: "Agent ID", value: "01J8522TYBZ98BQN1SGB61B1JY" },
    { label: "Connected", value: "Sep 19 2024 17:23" },
  ];

  return (
    <div className="pt-4 pl-4 pr-4 pb-0">
      <div className="mb-4">
        <Typography
          variant="h6"
          color="blue-gray"
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          Agents &gt; gallium
        </Typography>
      </div>

      <Card
        className="mb-4"
        placeholder={undefined}
        onPointerEnterCapture={undefined}
        onPointerLeaveCapture={undefined}
      >
        <CardBody
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          <Typography
            variant="h5"
            color="blue-gray"
            className="mb-4"
            placeholder={undefined}
            onPointerEnterCapture={undefined}
            onPointerLeaveCapture={undefined}
          >
            Agent Details
          </Typography>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3">
            {details.map((item, index) => (
              <div key={index}>
                <Typography
                  variant="small"
                  color="blue-gray"
                  className="font-medium"
                  placeholder={undefined}
                  onPointerEnterCapture={undefined}
                  onPointerLeaveCapture={undefined}
                >
                  {item.label}
                </Typography>
                <Typography
                  variant="small"
                  className="text-gray-700"
                  placeholder={undefined}
                  onPointerEnterCapture={undefined}
                  onPointerLeaveCapture={undefined}
                >
                  {item.value}
                </Typography>
              </div>
            ))}
          </div>
        </CardBody>
      </Card>
    </div>
  );
}

export function EditConfig() {
  return (
    <>
      <AgentDetails />
      <div className="flex justify-between items-center mt-4 pl-10 pb-4">
        <Typography
          className=""
          variant="h4"
          color="blue-gray"
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          Config
        </Typography>
        <div className="flex space-x-4 mr-10">
          {" "}
          {/* Flex container for the buttons with spacing */}
          <Button
            color="green"
            placeholder={undefined}
            onPointerEnterCapture={undefined}
            onPointerLeaveCapture={undefined}
          >
            Save
          </Button>
          <Button
            color="red"
            placeholder={undefined}
            onPointerEnterCapture={undefined}
            onPointerLeaveCapture={undefined}
          >
            Cancel
          </Button>
        </div>
      </div>
      <Editor
        height="55vh"
        defaultLanguage="javascript"
        defaultValue="// some comment"
      />
    </>
  );
}