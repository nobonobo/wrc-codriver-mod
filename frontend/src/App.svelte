<script>
  import { Events } from "@wailsio/runtime";
  import { Client } from "$bindings/wrc-codriver-mod/obs";
  import { Settings } from "$bindings/wrc-codriver-mod/settings";
  import { onMount } from "svelte";
  let auto_start_recording = false;
  let auto_stop_recording = false;
  let auto_youtube_upload = false;
  let stage_info = "";
  let vehicle_info = "";
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
    switch (ev.data.outputState) {
      case "OBS_WEBSOCKET_OUTPUT_STOPPED":
        stopDisabled = true;
    }
  });
  Events.On("recording", async (ev) => {
    stopDisabled = ev.data.recording;
    stage_info = ev.data.location + " / " + ev.data.route;
    vehicle_info =
      ev.data.class + " / " + ev.data.manufacturer + " / " + "vehicle";
    console.log(ev.data);
  });
  Events.On("finished", async (ev) => {
    console.log(ev.data);
  });
  Events.On("tyre-state", async (ev) => {
    console.log(ev.data);
  });
  Events.On("packet", async (ev) => {
    progress = ev.data.stage_progress * 100;
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
        <span class="label-text">Vehicle:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={vehicle_info}
          disabled
        />
      </label>
      <label class="label">
        <span class="label-text">Stage:</span>
        <input
          type="text"
          class="input input-sm input-ghost"
          value={stage_info}
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
