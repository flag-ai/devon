import { useQuery } from "@tanstack/react-query";
import { Link } from "react-router-dom";
import { getSetupStatus } from "../api/client";

export default function SetupBanner() {
  const { data } = useQuery({
    queryKey: ["setup-status"],
    queryFn: getSetupStatus,
  });

  if (!data || data.configured) return null;

  return (
    <div className="mb-6 rounded-lg border border-ctp-yellow/30 bg-ctp-yellow/5 px-4 py-3">
      <div className="flex items-center justify-between">
        <div>
          <p className="font-medium text-ctp-yellow">Setup recommended</p>
          <p className="text-sm text-ctp-subtext0 mt-0.5">
            Devon is running with defaults.{" "}
            {data.missing.length > 0 && (
              <>Configure {data.missing.join(", ")} for the best experience.</>
            )}
          </p>
        </div>
        <Link
          to="/settings"
          className="rounded-lg bg-ctp-yellow/10 px-4 py-2 text-sm font-medium text-ctp-yellow hover:bg-ctp-yellow/20 transition-colors"
        >
          Configure
        </Link>
      </div>
    </div>
  );
}
