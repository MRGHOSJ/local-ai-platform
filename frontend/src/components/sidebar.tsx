import { useLocation, useNavigate } from "react-router-dom";
import { cn } from "@/lib/utils";
import {
  LayoutDashboard,
  Bot,
  HardDrive,
  Users,
  Shield,
  Settings,
  Cpu,
} from "lucide-react";

const navItems = [
  { icon: LayoutDashboard, label: "Overview", path: "/" },
  { icon: Bot, label: "Agents", path: "/agents", disabled: true },
  { icon: HardDrive, label: "Models", path: "/models" },
  { icon: Users, label: "Friends", path: "/friends", disabled: true },
  { icon: Shield, label: "Security", path: "/security", disabled: true },
  { icon: Settings, label: "Settings", path: "/settings", disabled: true },
  { icon: Cpu, label: "System", path: "/system" },
];

export function Sidebar() {
  const location = useLocation();
  const navigate = useNavigate();

  return (
    <aside className="w-14 lg:w-48 h-full bg-card border-r flex flex-col py-4 shrink-0">
      <div className="px-3 mb-6">
        <h1 className="hidden lg:block text-lg font-bold bg-gradient-to-r from-primary to-secondary bg-clip-text text-transparent">
          Local AI
        </h1>
        <h1 className="lg:hidden text-lg font-bold text-primary">
          AI
        </h1>
      </div>
      <nav className="flex-1 space-y-1 px-2">
        {navItems.map((item) => (
          <button
            key={item.path}
            onClick={() => !item.disabled && navigate(item.path)}
            disabled={item.disabled}
            className={cn(
              "w-full flex items-center gap-3 px-2 py-2 rounded-md text-sm transition-colors",
              location.pathname === item.path
                ? "bg-primary/10 text-primary font-medium"
                : item.disabled
                ? "text-muted-foreground/50 cursor-not-allowed"
                : "text-muted-foreground hover:bg-accent hover:text-accent-foreground"
            )}
          >
            <item.icon className="h-4 w-4 shrink-0" />
            <span className="hidden lg:inline">{item.label}</span>
          </button>
        ))}
      </nav>
    </aside>
  );
}