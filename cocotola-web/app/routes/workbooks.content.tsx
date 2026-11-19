import { useTranslation } from "react-i18next";
import { isRouteErrorResponse, Link, Outlet, useLocation } from "react-router";
import { Button } from "~/components/ui/button";
import type { Route } from "./+types/workbooks.content";

export default function WorkbooksContent() {
  return <Outlet />;
}

type RouteErrorKind = "forbidden" | "notFound" | "conflict" | "server" | "unexpected";

function routeErrorKind(error: unknown): RouteErrorKind {
  if (!isRouteErrorResponse(error)) return "unexpected";
  if (error.status === 403) return "forbidden";
  if (error.status === 404) return "notFound";
  if (error.status === 409) return "conflict";
  if (error.status >= 500) return "server";
  return "unexpected";
}

function isReloadable(kind: RouteErrorKind): boolean {
  return kind === "conflict" || kind === "server";
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  const { t } = useTranslation();
  const location = useLocation();
  const kind = routeErrorKind(error);

  return (
    <div
      data-testid={`route-error-${kind}`}
      className="mx-auto flex max-w-2xl flex-col items-center gap-4 py-16 text-center"
    >
      <h1 className="text-2xl font-bold">{t(`routeErrors.${kind}.title`)}</h1>
      <p className="text-muted-foreground">{t(`routeErrors.${kind}.description`)}</p>
      <div className="flex flex-col items-center gap-3 sm:flex-row">
        {isReloadable(kind) && (
          <Button
            variant="outline"
            nativeButton={false}
            render={
              <Link
                reloadDocument
                to={`${location.pathname}${location.search}`}
                data-testid="route-error-reload"
              />
            }
          >
            {t("routeErrors.reload")}
          </Button>
        )}
        <Button
          nativeButton={false}
          render={<Link to="/workbooks" data-testid="route-error-back" />}
        >
          {t("routeErrors.backToWorkbooks")}
        </Button>
      </div>
    </div>
  );
}
