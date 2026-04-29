import { index, layout, type RouteConfig, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("login", "routes/login.tsx"),
  route("signup", "routes/signup.tsx"),
  route("auth/callback", "routes/auth.callback.tsx"),
  route("logout", "routes/logout.tsx"),
  route("user-language", "routes/user-language.tsx"),
  layout("routes/workbooks.tsx", [
    route("workbooks", "routes/workbooks.index.tsx"),
    route("workbooks/public", "routes/workbooks.public.tsx"),
    route("workbooks/:workbookId", "routes/workbooks.$workbookId.tsx"),
    route("workbooks/:workbookId/study", "routes/workbooks.$workbookId.study.tsx"),
  ]),
] satisfies RouteConfig;
