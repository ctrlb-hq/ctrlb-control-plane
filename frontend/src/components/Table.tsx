import { PencilIcon } from "@heroicons/react/24/solid";
import { MagnifyingGlassIcon } from "@heroicons/react/24/outline";
// import queryService from "../services/queryServices";
import {
  Card,
  Typography,
  Button,
  CardBody,
  Chip,
  CardFooter,
  CardHeader,
  Avatar,
  IconButton,
  Tooltip,
  Tabs,
  Tab,
  TabsHeader,
  Input,
} from "@material-tailwind/react";

import { useNavigate } from "react-router-dom";
import authService from "../services/authService";

const TABS = [
  {
    label: "Pipelines",
    value: "pipelines",
  },
  {
    label: "Agents",
    value: "agents",
  },
];

const TABLE_HEAD = ["Name", "Type", "Status", "Exported Volume", ""];

const TABLE_ROWS = [
  {
    img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg-1.jpg",
    name: "Agent Alpha",
    type: "Linux",
    version: "v2.0.1",
    status: "Active",
    exportedVolume: "150 GB",
    online: true,
    date: "12/03/22",
  },
  {
    img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg",
    name: "Agent Beta",
    type: "Windows",
    version: "v1.3.5",
    status: "Inactive",
    exportedVolume: "85 GB",
    online: false,
    date: "28/06/21",
  },
  {
    img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg",
    name: "Agent Gamma",
    type: "MacOS",
    version: "v1.2.0",
    status: "Active",
    exportedVolume: "200 GB",
    online: true,
    date: "15/09/23",
  },
  {
    img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg",
    name: "Agent Delta",
    type: "Linux",
    version: "v3.0.0",
    status: "Active",
    exportedVolume: "320 GB",
    online: true,
    date: "03/11/23",
  },
  {
    img: "https://cdn.brandfetch.io/idxVhszl6V/w/400/h/400/theme/dark/icon.jpeg",
    name: "Agent Epsilon",
    type: "Windows",
    version: "v1.0.0",
    status: "Inactive",
    exportedVolume: "50 GB",
    online: false,
    date: "21/07/20",
  },
];

function AgentsTable() {
  const navigate = useNavigate();

  const handleClick = () => {
    navigate("/config/123"); // Redirects to /config/123
  };
  return (
    <table className="mt-4 w-full min-w-auto table-fixed text-left ">
      <thead>
        <tr>
          {TABLE_HEAD.map((head) => (
            <th
              key={head}
              className="border-y border-blue-gray-100 bg-blue-gray-50/50 p-4"
            >
              <Typography
                variant="small"
                color="blue-gray"
                className="font-normal leading-none opacity-70"
                placeholder={undefined}
                onPointerEnterCapture={undefined}
                onPointerLeaveCapture={undefined}
              >
                {head}
              </Typography>
            </th>
          ))}
        </tr>
      </thead>
      <tbody>
        {TABLE_ROWS.map(
          ({ img, name, type, version, status, exportedVolume }, index) => {
            const isLast = index === TABLE_ROWS.length - 1;
            const classes = isLast ? "p-4" : "p-4 border-b border-blue-gray-50";

            return (
              <tr key={name}>
                <td className={classes}>
                  <div className="flex items-center gap-3">
                    <Avatar
                      src={img}
                      alt={name}
                      size="sm"
                      placeholder={"NA"}
                      onPointerEnterCapture={undefined}
                      onPointerLeaveCapture={undefined}
                    />
                    <div className="flex flex-col">
                      <Typography
                        variant="small"
                        color="blue-gray"
                        className="font-normal"
                        placeholder={undefined}
                        onPointerEnterCapture={undefined}
                        onPointerLeaveCapture={undefined}
                      >
                        {name}
                      </Typography>
                    </div>
                  </div>
                </td>
                <td className={classes}>
                  <div className="flex flex-col">
                    <Typography
                      variant="small"
                      color="blue-gray"
                      className="font-normal"
                      placeholder={undefined}
                      onPointerEnterCapture={undefined}
                      onPointerLeaveCapture={undefined}
                    >
                      {type}
                    </Typography>
                    <Typography
                      variant="small"
                      color="blue-gray"
                      className="font-normal opacity-70"
                      placeholder={undefined}
                      onPointerEnterCapture={undefined}
                      onPointerLeaveCapture={undefined}
                    >
                      {version}
                    </Typography>
                  </div>
                </td>
                <td className={classes}>
                  <div className="w-max">
                    <Chip
                      variant="ghost"
                      size="sm"
                      value={status}
                      color={status == "Active" ? "green" : "red"}
                    />
                  </div>
                </td>
                <td className={classes}>
                  <Typography
                    variant="small"
                    color="blue-gray"
                    className="font-normal"
                    placeholder={undefined}
                    onPointerEnterCapture={undefined}
                    onPointerLeaveCapture={undefined}
                  >
                    {exportedVolume}
                  </Typography>
                </td>
                <td className={classes}>
                  <Tooltip content="Edit Config">
                    <IconButton
                      variant="text"
                      placeholder={undefined}
                      onPointerEnterCapture={undefined}
                      onPointerLeaveCapture={undefined}
                      onClick={handleClick}
                    >
                      <PencilIcon className="h-4 w-4" />
                    </IconButton>
                  </Tooltip>
                </td>
              </tr>
            );
          }
        )}
      </tbody>
    </table>
  );
}

function Pagination() {
  return (
    <>
      <Typography
        variant="small"
        color="blue-gray"
        className="font-normal"
        placeholder={undefined}
        onPointerEnterCapture={undefined}
        onPointerLeaveCapture={undefined}
      >
        Page 1 of 10
      </Typography>
      <div className="flex gap-2">
        <Button
          variant="outlined"
          size="sm"
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          Previous
        </Button>
        <Button
          variant="outlined"
          size="sm"
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          Next
        </Button>
      </div>
    </>
  );
}

export function MembersTable() {
  const navigate = useNavigate();

  // queryService.fetchAgents()
  // queryService.fetchPipelines()

  const handleLogout = async () => {
    try {
      await authService.logout();
      navigate('/login');
    } catch (error) {
      console.error('Logout failed:', error);
    }
  };

  return (
    <Card
      className="h-full w-full"
      placeholder={undefined}
      onPointerEnterCapture={undefined}
      onPointerLeaveCapture={undefined}
    >
      <CardHeader
        floated={false}
        shadow={false}
        className="rounded-none"
        placeholder={undefined}
        onPointerEnterCapture={undefined}
        onPointerLeaveCapture={undefined}
      >
        <div className="flex flex-col items-center justify-between gap-4 md:flex-row mt-4">
          <div className="flex items-center justify-between w-full">
            <Tabs value="all" className="w-full md:w-max">
              <TabsHeader
                placeholder={undefined}
                onPointerEnterCapture={undefined}
                onPointerLeaveCapture={undefined}
              >
                {TABS.map(({ label, value }) => (
                  <Tab
                    key={value}
                    value={value}
                    placeholder={undefined}
                    onPointerEnterCapture={undefined}
                    onPointerLeaveCapture={undefined}
                  >
                    &nbsp;&nbsp;{label}&nbsp;&nbsp;
                  </Tab>
                ))}
              </TabsHeader>
            </Tabs>
            
            <div className="flex items-center gap-4">
              <div className="w-72">
                <Input
                  label="Search"
                  icon={<MagnifyingGlassIcon className="h-5 w-5" />}
                  onPointerEnterCapture={undefined}
                  onPointerLeaveCapture={undefined}
                  crossOrigin={undefined}
                />
              </div>
              <Button
                size="sm"
                variant="outlined"
                className="flex items-center gap-2 text-red-500 border-red-500 hover:bg-red-50"
                onClick={handleLogout}
                placeholder={undefined}
                onPointerEnterCapture={undefined}
                onPointerLeaveCapture={undefined}
              >
                <svg
                  xmlns="http://www.w3.org/2000/svg"
                  fill="none"
                  viewBox="0 0 24 24"
                  strokeWidth={2}
                  stroke="currentColor"
                  className="h-4 w-4"
                >
                  <path
                    strokeLinecap="round"
                    strokeLinejoin="round"
                    d="M5.636 5.636a9 9 0 1012.728 0M12 3v9"
                  />
                </svg>
                Logout
              </Button>
            </div>
          </div>
        </div>
      </CardHeader>
      <CardBody
        className="overflow-scroll px-0 p-4"
        placeholder={undefined}
        onPointerEnterCapture={undefined}
        onPointerLeaveCapture={undefined}
      >
        <AgentsTable />
      </CardBody>
      {TABS.length > 10 && (
        <CardFooter
          className="flex items-center justify-between border-t border-blue-gray-50 p-4"
          placeholder={undefined}
          onPointerEnterCapture={undefined}
          onPointerLeaveCapture={undefined}
        >
          <Pagination />
        </CardFooter>
      )}
    </Card>
  );
}
