<script>
    import "./scss/main.scss";

    import Router, { replace, link } from "svelte-spa-router";
    import active from "svelte-spa-router/active";
    import routes from "./routes";
    import ApiClient from "@/utils/ApiClient";
    import CommonHelper from "@/utils/CommonHelper";
    import tooltip from "@/actions/tooltip";
    import Toasts from "@/components/base/Toasts.svelte";
    import Toggler from "@/components/base/Toggler.svelte";
    import Confirmation from "@/components/base/Confirmation.svelte";
    import { admin } from "@/stores/admin";
    import { setErrors } from "@/stores/errors";
    import { resetConfirmation } from "@/stores/confirmation";

    let oldLocation = undefined;

    function handleRouteLoading(e) {
        if (e?.detail?.location === oldLocation) {
            return; // not an actual change
        }

        oldLocation = e?.detail?.location;

        // resets
        CommonHelper.setDocumentTitle("");
        setErrors({});
        resetConfirmation();
    }

    function handleRouteFailure() {
        if (ApiClient.AuthStore.isValid) {
            replace("/");
        } else {
            ApiClient.logout();
        }
    }

    function logout() {
        ApiClient.logout();
    }
</script>

<div class="app-layout">
    {#if $admin?.id}
        <aside class="app-sidebar">
            <a href="/" class="logo logo-sm" use:link>
                <img
                    src="{import.meta.env.BASE_URL}images/logo.svg"
                    alt="PocketBase logo"
                    width="40"
                    height="40"
                />
            </a>

            <nav class="main-menu">
                <a
                    href="/collections"
                    class="menu-item"
                    aria-label="Collections"
                    use:link
                    use:active={{ path: "/collections/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Collections", position: "right" }}
                >
                    <i class="ri-database-2-line" />
                </a>
                <a
                    href="/users"
                    class="menu-item"
                    aria-label="Users"
                    use:link
                    use:active={{ path: "/users/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Users", position: "right" }}
                >
                    <i class="ri-group-line" />
                </a>
                <a
                    href="/logs"
                    class="menu-item"
                    aria-label="Logs"
                    use:link
                    use:active={{ path: "/logs/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Logs", position: "right" }}
                >
                    <i class="ri-line-chart-line" />
                </a>
                <a
                    href="/settings"
                    class="menu-item"
                    aria-label="Settings"
                    use:link
                    use:active={{ path: "/settings/?.*", className: "current-route" }}
                    use:tooltip={{ text: "Settings", position: "right" }}
                >
                    <i class="ri-tools-line" />
                </a>
            </nav>

            <figure class="thumb thumb-circle link-hint closable">
                <img
                    src="{import.meta.env.BASE_URL}images/avatars/avatar{$admin?.avatar || 0}.svg"
                    alt="Avatar"
                />
                <Toggler class="dropdown dropdown-nowrap dropdown-upside dropdown-left">
                    <a href="/settings/admins" class="dropdown-item closable" use:link>
                        <i class="ri-shield-user-line" />
                        <span class="txt">Manage admins</span>
                    </a>
                    <hr />
                    <div tabindex="0" class="dropdown-item closable" on:click={logout}>
                        <i class="ri-logout-circle-line" />
                        <span class="txt">Logout</span>
                    </div>
                </Toggler>
            </figure>
        </aside>
    {/if}

    <div class="app-body">
        <Router {routes} on:routeLoading={handleRouteLoading} on:conditionsFailed={handleRouteFailure} />
    </div>
</div>

<Toasts />

<Confirmation />
