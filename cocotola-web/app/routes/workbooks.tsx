import { ActivityIcon, BookOpenIcon, GlobeIcon, LogOutIcon } from "lucide-react";
import { useTranslation } from "react-i18next";
import { Form, Link, Outlet, useFetcher, useLoaderData, useLocation } from "react-router";
import { ConfirmDialogProvider } from "~/components/confirm-dialog-provider";
import { Button } from "~/components/ui/button";
import { Separator } from "~/components/ui/separator";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarInset,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarProvider,
  SidebarSeparator,
  SidebarTrigger,
} from "~/components/ui/sidebar";
import { type SupportedLanguage, supportedLanguages } from "~/i18n/config";
import { getUserPreferences, type UserPreferences } from "~/lib/api/user-setting.server";
import { requireAuth } from "~/lib/auth/require-auth.server";
import type { Route } from "./+types/workbooks";

export async function loader({ request }: Route.LoaderArgs) {
  console.info("[workbooks] loader called");
  const { accessToken } = await requireAuth(request);

  // Layout-level loader: render the outlet even when /auth/me transiently
  // fails (5xx) so the rest of the app stays reachable. The original
  // implementation returned `{ user: null }` on non-401 errors; that
  // graceful degradation was lost when this loader was refactored onto
  // getUserPreferences, which throws Response for every non-2xx. Catch
  // those Responses here and fall back to null — but propagate the 3xx
  // Responses (the redirect("/login") that redirectOnUnauthorized fires
  // on 401) so session expiry still takes the user to the login screen.
  let user: UserPreferences | null = null;
  try {
    user = await getUserPreferences(request, accessToken);
  } catch (e) {
    if (e instanceof Response && e.status >= 300 && e.status < 400) throw e;
    console.error("[workbooks] /auth/me failed (degraded render):", e);
  }

  if (user !== null) {
    // loginId is email-shaped PII; log only the opaque userId so server
    // logs stay compliance-safe. Richer per-request attributes belong on
    // OTEL spans, not in console output.
    console.info(`[workbooks] user loaded: userId=${user.userId}`);
  }

  return { user };
}

const navItems = [
  {
    titleKey: "workbooks.nav.dashboard",
    href: "/dashboard",
    icon: ActivityIcon,
    disabled: false,
  },
  {
    titleKey: "workbooks.nav.myWorkbooks",
    href: "/workbooks",
    icon: BookOpenIcon,
    disabled: false,
  },
  {
    titleKey: "workbooks.nav.public",
    href: "/workbooks/public",
    icon: GlobeIcon,
    disabled: false,
  },
];

export default function WorkbooksLayout() {
  const { user } = useLoaderData<typeof loader>();
  const location = useLocation();
  const { t, i18n } = useTranslation();
  const languageFetcher = useFetcher();
  const isChangingLanguage = languageFetcher.state !== "idle";

  function handleLanguageChange(event: React.ChangeEvent<HTMLSelectElement>) {
    const nextLanguage = event.target.value as SupportedLanguage;
    if (nextLanguage === i18n.language) return;
    i18n.changeLanguage(nextLanguage);
    languageFetcher.submit(
      { language: nextLanguage },
      { method: "post", action: "/user-language" },
    );
  }

  return (
    <ConfirmDialogProvider>
      <SidebarProvider>
        <Sidebar>
          <SidebarHeader>
            <Link to="/workbooks" className="flex items-center gap-2 px-2 py-1">
              <BookOpenIcon className="size-5" />
              <span className="text-lg font-bold">Cocotola</span>
            </Link>
          </SidebarHeader>
          <SidebarSeparator />
          <SidebarContent>
            <SidebarGroup>
              <SidebarGroupLabel>{t("workbooks.nav.sidebarLabel")}</SidebarGroupLabel>
              <SidebarGroupContent>
                <SidebarMenu>
                  {navItems.map((item) => {
                    const title = t(item.titleKey);
                    return (
                      <SidebarMenuItem key={item.titleKey}>
                        <SidebarMenuButton
                          isActive={!item.disabled && location.pathname === item.href}
                          render={
                            item.disabled ? (
                              <span className="opacity-50 cursor-not-allowed" />
                            ) : (
                              <Link to={item.href} />
                            )
                          }
                          tooltip={
                            item.disabled ? `${title} (${t("workbooks.nav.comingSoon")})` : title
                          }
                        >
                          <item.icon className="size-4" />
                          <span>{title}</span>
                          {item.disabled && (
                            <span className="ml-auto rounded-full bg-muted px-2 py-0.5 text-[10px] text-muted-foreground">
                              {t("workbooks.nav.comingSoon")}
                            </span>
                          )}
                        </SidebarMenuButton>
                      </SidebarMenuItem>
                    );
                  })}
                </SidebarMenu>
              </SidebarGroupContent>
            </SidebarGroup>
          </SidebarContent>
          <SidebarSeparator />
          <SidebarFooter>
            {user && (
              <div className="flex items-center justify-between gap-2 px-2">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{user.loginId}</p>
                  <p className="truncate text-xs text-muted-foreground">{user.organizationName}</p>
                </div>
                <div className="flex items-center gap-1">
                  <select
                    value={i18n.language}
                    onChange={handleLanguageChange}
                    disabled={isChangingLanguage}
                    aria-label={t("workbooks.nav.languageLabel")}
                    aria-busy={isChangingLanguage}
                    className="h-8 rounded-md border border-input bg-transparent px-2 text-xs"
                  >
                    {supportedLanguages.map((lang) => (
                      <option key={lang} value={lang}>
                        {t(`languages.${lang}`, { defaultValue: lang.toUpperCase() })}
                      </option>
                    ))}
                  </select>
                  <Form method="post" action="/logout">
                    <Button variant="ghost" size="icon-sm" type="submit">
                      <LogOutIcon className="size-4" />
                      <span className="sr-only">{t("workbooks.nav.logout")}</span>
                    </Button>
                  </Form>
                </div>
              </div>
            )}
          </SidebarFooter>
        </Sidebar>
        <SidebarInset>
          <header className="flex h-12 items-center gap-2 border-b px-4">
            <SidebarTrigger />
            <Separator orientation="vertical" className="mx-1 h-4" />
          </header>
          <div className="flex-1 overflow-auto p-6">
            <Outlet />
          </div>
        </SidebarInset>
      </SidebarProvider>
    </ConfirmDialogProvider>
  );
}
