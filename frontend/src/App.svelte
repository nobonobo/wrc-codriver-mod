<script>
  import { Events } from "@wailsio/runtime";
  import { Client } from "$bindings/wrc-codriver-mod/obs";
  import { Settings } from "$bindings/wrc-codriver-mod/settings";
  import { onMount } from "svelte";
  let auto_start_recording = false;
  let auto_stop_recording = false;
  let auto_youtube_upload = false;
  let location = "";
  let route = "";
  let category = "";
  let manufacturer = "";
  let vehicle = "";
  let progress = 0.0;
  let result = "";
  let stopDisabled = true;
  async function onChange(ev) {
    await Settings.Update({
      auto_start_recording,
      auto_stop_recording,
      auto_youtube_upload,
    });
  }
  async function stopRecord() {
    await Client.StopRecord();
    stopDisabled = true;
  }
  onMount(async () => {
    console.log("start");
    let settings = await Settings.Get();
    auto_start_recording = settings.auto_start_recording;
    auto_stop_recording = settings.auto_stop_recording;
    auto_youtube_upload = settings.auto_youtube_upload;
    Events.On("obs-event", async (ev) => {
      switch (ev.data[0].outputState) {
        case "OBS_WEBSOCKET_OUTPUT_STOPPED":
          stopDisabled = true;
      }
    });
    Events.On("recording", async (ev) => {
      console.log(ev.data[0]);
      stopDisabled = ev.data[0].recording;
      location = ev.data[0].location;
      route = ev.data[0].route;
      category = ev.data[0].class;
      manufacturer = ev.data[0].manufacturer;
      vehicle = ev.data[0].vehicle;
    });
    Events.On("finished", async (ev) => {
      console.log(ev.data[0]);
    });
    Events.On("tyre-state", async (ev) => {
      console.log(ev.data[0]);
    });
    Events.On("packet", async (ev) => {
      progress = ev.data[0].stage_progress * 100;
    });
    Events.On("result", async (ev) => {
      result = ev.data[0];
    });
  });
</script>

<div class="container">
  <div class="card mx-8">
    <label class="label">
      <span>Auto Start Recording:</span>
      <input
        type="checkbox"
        class="toggle"
        bind:checked={auto_start_recording}
        on:change={onChange}
      />
    </label>
    <label class="label">
      <span>Auto Stop Recording:</span>
      <input
        type="checkbox"
        class="toggle"
        bind:checked={auto_stop_recording}
        on:change={onChange}
      />
    </label>
    <label class="label">
      <span>Recording:</span>
      <div class="join">
        <button
          class="btn btn-sm btn-primary join-item"
          on:click={async () => await Client.StopRecord()}
          disabled={!stopDisabled}>Start</button
        >
        <button
          class="btn btn-sm btn-secondary join-item"
          on:click={stopRecord}
          disabled={stopDisabled}>Stop</button
        >
      </div>
    </label>
    <label class="label">
      <span>Auto Upload To YouTube:</span>
      <input
        type="checkbox"
        class="toggle"
        bind:checked={auto_youtube_upload}
        on:change={onChange}
      />
    </label>
    <label class="label block">
      <span>Class:</span>
      <input
        type="text"
        class="input input-sm block w-full"
        value={category}
        disabled
      />
    </label>
    <label class="label block">
      <span>Manufacturer:</span>
      <input
        type="text"
        class="input input-sm block w-full"
        value={manufacturer}
        disabled
      />
    </label>
    <label class="label block">
      <span>Vehicle:</span>
      <input
        type="text"
        class="input input-sm block w-full"
        value={vehicle}
        disabled
      />
    </label>
    <label class="label block">
      <span>Location:</span>
      <input
        type="text"
        class="input input-sm block w-full"
        value={location}
        disabled
      />
    </label>
    <label class="label block">
      <span>Route:</span>
      <input
        type="text"
        class="input input-sm block w-full"
        value={route}
        disabled
      />
    </label>
    <label class="label block">
      <span>Progress:</span>
      <progress
        class="progress progress-primary w-full"
        value={progress}
        max="100"
      ></progress>
    </label>
    <label class="label block">
      <span>Result:</span>
      <textarea
        class="textarea block w-full"
        rows="2"
        disabled
        bind:value={result}
      />
    </label>
  </div>
</div>

<style>
  /* Put your standard CSS here */
</style>
