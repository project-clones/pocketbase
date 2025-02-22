<script>
    import {
        collections,
        activeCollection,
        isCollectionsLoading,
        loadCollections,
    } from "@/stores/collections";
    import CommonHelper from "@/utils/CommonHelper";
    import tooltip from "@/actions/tooltip";
    import Searchbar from "@/components/base/Searchbar.svelte";
    import CollectionsSidebar from "@/components/collections/CollectionsSidebar.svelte";
    import CollectionUpsertPanel from "@/components/collections/CollectionUpsertPanel.svelte";
    import CollectionDocsPanel from "@/components/collections/docs/CollectionDocsPanel.svelte";
    import RecordUpsertPanel from "@/components/records/RecordUpsertPanel.svelte";
    import RecordsList from "@/components/records/RecordsList.svelte";

    const queryParams = CommonHelper.getQueryParams(window.location?.href);

    let collectionUpsertPanel;
    let collectionDocsPanel;
    let recordPanel;
    let recordsList;
    let filter = queryParams.filter || "";
    let sort = queryParams.sort || "-created";
    let selectedCollectionId = queryParams.collectionId;

    $: viewableCollections = $collections.filter((c) => c.name != import.meta.env.PB_PROFILE_COLLECTION);

    // reset filter and sort on collection change
    $: if ($activeCollection?.id && selectedCollectionId != $activeCollection.id) {
        selectedCollectionId = $activeCollection.id;
        sort = "-created";
        filter = "";
    }

    // keep the url params in sync
    $: if (sort || filter || $activeCollection?.id) {
        CommonHelper.replaceClientQueryParams({
            collectionId: $activeCollection?.id,
            filter: filter,
            sort: sort,
        });
    }

    CommonHelper.setDocumentTitle("Collections");

    loadCollections(selectedCollectionId);
</script>

{#if $isCollectionsLoading}
    <div class="placeholder-section m-b-base">
        <span class="loader loader-lg" />
        <h1>Loading collections...</h1>
    </div>
{:else if !viewableCollections.length}
    <div class="placeholder-section m-b-base">
        <div class="icon">
            <i class="ri-database-2-line" />
        </div>
        <h1 class="m-b-10">Create your first collection to add records!</h1>
        <button
            type="button"
            class="btn btn-expanded-lg btn-lg"
            on:click={() => collectionUpsertPanel?.show()}
        >
            <i class="ri-add-line" />
            <span class="txt">Create new collection</span>
        </button>
    </div>
{:else}
    <CollectionsSidebar />

    <main class="page-wrapper">
        <header class="page-header">
            <nav class="breadcrumbs">
                <div class="breadcrumb-item">Collections</div>
                <div class="breadcrumb-item">{$activeCollection.name}</div>
            </nav>

            <button
                type="button"
                class="btn btn-secondary btn-circle"
                use:tooltip={{ text: "Edit collection", position: "right" }}
                on:click={() => collectionUpsertPanel?.show($activeCollection)}
            >
                <i class="ri-settings-4-line" />
            </button>

            <div class="btns-group">
                <button
                    type="button"
                    class="btn btn-outline"
                    on:click={() => collectionDocsPanel?.show($activeCollection)}
                >
                    <i class="ri-code-s-slash-line" />
                    <span class="txt">API Preview</span>
                </button>

                <button type="button" class="btn btn-expanded" on:click={() => recordPanel?.show()}>
                    <i class="ri-add-line" />
                    <span class="txt">New record</span>
                </button>
            </div>
        </header>

        <Searchbar
            value={filter}
            autocompleteCollection={$activeCollection}
            on:submit={(e) => (filter = e.detail)}
        />

        <RecordsList
            bind:this={recordsList}
            collection={$activeCollection}
            bind:filter
            bind:sort
            on:select={(e) => recordPanel?.show(e?.detail)}
        />
    </main>
{/if}

<CollectionUpsertPanel bind:this={collectionUpsertPanel} />
<CollectionDocsPanel bind:this={collectionDocsPanel} />

<RecordUpsertPanel
    bind:this={recordPanel}
    collection={$activeCollection}
    on:save={() => recordsList?.load()}
    on:delete={() => recordsList?.load()}
/>
