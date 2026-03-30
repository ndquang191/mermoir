import { createRootRoute, Link, Outlet } from "@tanstack/react-router";

export const Route = createRootRoute({
  component: RootLayout,
});

function RootLayout() {
  return (
    <div style={{ fontFamily: "system-ui, sans-serif", maxWidth: 800, margin: "0 auto", padding: "0 1rem" }}>
      <header style={{ borderBottom: "1px solid #eee", padding: "1rem 0", marginBottom: "2rem" }}>
        <h1 style={{ margin: 0, display: "inline", fontSize: "1.5rem" }}>Memoir</h1>
        <nav style={{ display: "inline", marginLeft: "2rem" }}>
          <Link
            to="/"
            style={{ marginRight: "1rem", textDecoration: "none", color: "#333" }}
            activeProps={{ style: { fontWeight: "bold" } }}
          >
            Home
          </Link>
          <Link
            to="/new"
            style={{ textDecoration: "none", color: "#333" }}
            activeProps={{ style: { fontWeight: "bold" } }}
          >
            New Memory
          </Link>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
}
