import { useTranslation } from "react-i18next";
import { Link } from "react-router";
import { Button } from "~/components/ui/button";
import { classifyRouteError, type RouteErrorKind } from "~/lib/route-error";

const MY_WORKBOOKS_PATH = "/workbooks";

const RELOADABLE_KINDS: ReadonlySet<RouteErrorKind> = new Set(["conflict", "server", "unexpected"]);

export function RouteErrorView({ error }: { error: unknown }) {
  const { t } = useTranslation();
  const kind = classifyRouteError(error);
  const devStack = import.meta.env.DEV && error instanceof Error ? error.stack : undefined;

  return (
    <div
      data-testid={`route-error-${kind}`}
      className="mx-auto flex max-w-xl flex-col items-center gap-4 py-16 text-center"
    >
      <h1 className="text-2xl font-bold">{t(`routeErrors.${kind}.title`)}</h1>
      <p className="text-muted-foreground">{t(`routeErrors.${kind}.description`)}</p>
      {devStack && (
        <pre
          data-testid="route-error-stack"
          className="w-full overflow-x-auto rounded-md bg-muted p-4 text-left text-xs"
        >
          <code>{devStack}</code>
        </pre>
      )}
      <div className="flex flex-col items-center gap-3 sm:flex-row">
        {RELOADABLE_KINDS.has(kind) && (
          <Button data-testid="route-error-reload" onClick={() => window.location.reload()}>
            {t("routeErrors.reload")}
          </Button>
        )}
        <Button
          variant="outline"
          nativeButton={false}
          data-testid="route-error-back"
          render={<Link to={MY_WORKBOOKS_PATH} />}
        >
          {t("routeErrors.backToWorkbooks")}
        </Button>
      </div>
    </div>
  );
}
