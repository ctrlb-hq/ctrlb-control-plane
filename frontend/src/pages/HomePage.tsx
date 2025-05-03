import { useNavigate } from "react-router-dom";
import authService from "../services/authService";
import Pipeline from "../components/Pipelines/Pipeline";
import { ROUTES } from "../constants/routes";
import AddPipeline from "../components/Pipelines/AddPipeline";
import { Button } from "../components/ui/button";
import { ArrowLeftRight } from "lucide-react";

const TABS = [
  { label: "Pipelines", value: "pipelines", icon: <ArrowLeftRight /> },
];

export function HomePage() {
  const navigate = useNavigate();
  const handleLogout = async () => {
    try {
      await authService.logout();
      navigate(ROUTES.LOGIN, { replace: true });
    } catch (error) {
      console.error("Logout failed:", error);
      localStorage.removeItem('authToken');
      navigate(ROUTES.LOGIN, { replace: true });
    }
  };

  return (
    <div className="w-full h-full">
      <div className="p-4">
        <div className="flex flex-col mx-4 md:flex-row items-center justify-between gap-4">
          <div className="flex items-center w-full md:w-auto">
            <div className="flex gap-2 border-b">
                {TABS.map(({ label, value, icon }) => (
                <div key={value} className="flex items-center">
                  <button
                  className={`px-4 py-2 rounded-t-md text-gray-600`}
                  >
                  <span className="flex items-center gap-2">
                    {icon}
                    {label}
                  </span>
                  </button>
                </div>
                ))}
            </div>
          </div>

          <div className="flex items-center gap-2">
              <AddPipeline />
            <Button className="flex items-center gap-1 px-2 py-1" variant={"destructive"} onClick={handleLogout}>
              <svg
                xmlns="http://www.w3.org/2000/svg"
                fill="none"
                viewBox="0 0 24 24"
                stroke="currentColor"
                className="h-4 w-4"
              >
                <path
                  strokeLinecap="round"
                  strokeLinejoin="round"
                  strokeWidth={2}
                  d="M5.636 5.636a9 9 0 1012.728 0M12 3v9"
                />
              </svg>
              Logout
            </Button>
          </div>
        </div>
        {(
          <div className="p-4 rounded-md">
            <Pipeline />
          </div>
        )}
      </div>
    </div>
  );
}


export default HomePage;