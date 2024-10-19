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
  async function onChange(ev) {
    await Settings.Update({
      auto_start_recording,
      auto_stop_recording,
      auto_youtube_upload,
    });
  }
  onMount(async () => {
    let settings = await Settings.Get();
    auto_start_recording = settings.auto_start_recording;
    auto_stop_recording = settings.auto_stop_recording;
    auto_youtube_upload = settings.auto_youtube_upload;
  });
  let stopDisabled = true;
  let time = "Listening for Time event...";
  async function stopRecord() {
    await Client.StopRecord();
    stopDisabled = true;
  }
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
</script>

<div class="container">
  <div class="card mx-8">
    <div class="input-box">
      <label class="label cursor-pointer">
        <span class="label-text">Auto Start Recording:</span>
        <input
          type="checkbox"
          class="toggle"
          bind:checked={auto_start_recording}
          on:change={onChange}
        />
      </label>
      <label class="label cursor-pointer">
        <span class="label-text">Auto Stop Recording:</span>
        <input
          type="checkbox"
          class="toggle"
          bind:checked={auto_stop_recording}
          on:change={onChange}
        />
      </label>
      <label class="label">
        <span class="label-text">Recording:</span>
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
      <label class="label cursor-pointer">
        <span class="label-text">Auto Upload To YouTube:</span>
        <input
          type="checkbox"
          class="toggle"
          bind:checked={auto_youtube_upload}
          on:change={onChange}
        />
      </label>
      <label class="label">
        <span class="label-text">Class:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={category}
          disabled
        />
      </label>
      <label class="label">
        <span class="label-text">Manufacturer:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={manufacturer}
          disabled
        />
      </label>
      <label class="label">
        <span class="label-text">Vehicle:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={vehicle}
          disabled
        />
      </label>
      <label class="label">
        <span class="label-text">Location:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={location}
          disabled
        />
      </label>
      <label class="label">
        <span class="label-text">Route:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={route}
          disabled
        />
      </label>
      <label class="label cursor-pointer">
        <span class="label-text">Progress:</span>
        <progress
          class="progress progress-primary w-56"
          value={progress}
          max="100"
        ></progress>
      </label>
    </div>
  </div>
</div>

<style>
  /* Put your standard CSS here */
</style>
