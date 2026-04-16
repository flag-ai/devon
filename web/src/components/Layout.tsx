import { NavLink, Outlet } from "react-router-dom";
import { clearToken } from "../api";

const navLink =
  "block rounded px-3 py-2 text-sm hover:bg-mocha-surface0 text-mocha-subtext1";
const activeNavLink = "bg-mocha-surface0 text-mocha-lavender";

export default function Layout() {
  return (
    <div className="flex min-h-screen bg-mocha-base text-mocha-text">
      <aside className="w-56 border-r border-mocha-surface1 bg-mocha-mantle p-4">
        <h1 className="mb-6 text-xl font-semibold text-mocha-lavender">DEVON</h1>
        <nav className="space-y-1">
          <NavLink end to="/" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Dashboard
          </NavLink>
          <NavLink to="/search" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Search
          </NavLink>
          <NavLink to="/models" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Models
          </NavLink>
          <NavLink to="/downloads" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Downloads
          </NavLink>
          <NavLink to="/agents" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Agents
          </NavLink>
          <NavLink to="/settings" className={({ isActive }) => `${navLink} ${isActive ? activeNavLink : ""}`}>
            Settings
          </NavLink>
        </nav>
        <button
          type="button"
          onClick={() => {
            clearToken();
            window.location.reload();
          }}
          className="mt-8 text-xs text-mocha-overlay1 hover:text-mocha-red"
        >
          Sign out
        </button>
      </aside>
      <main className="flex-1 p-8">
        <Outlet />
      </main>
    </div>
  );
}
