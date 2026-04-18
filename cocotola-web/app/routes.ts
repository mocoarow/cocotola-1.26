import { index, layout, type RouteConfig, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("login", "routes/login.tsx"),
  route("auth/callback", "routes/auth.callback.tsx"),
  route("logout", "routes/logout.tsx"),
  layout("routes/workbooks.tsx", [
    route("workbooks", "routes/workbooks.index.tsx"),
    route("workbooks/:workbookId", "routes/workbooks.$workbookId.tsx"),
  ]),
] satisfies RouteConfig;
