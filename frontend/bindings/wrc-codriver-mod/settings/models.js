// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import {Create as $Create} from "@wailsio/runtime";

export class Info {
    /**
     * Creates a new Info instance.
     * @param {Partial<Info>} [$$source = {}] - The source object to create the Info.
     */
    constructor($$source = {}) {
        if (!("auto_start_recording" in $$source)) {
            /**
             * @member
             * @type {boolean}
             */
            this["auto_start_recording"] = false;
        }
        if (!("auto_stop_recording" in $$source)) {
            /**
             * @member
             * @type {boolean}
             */
            this["auto_stop_recording"] = false;
        }
        if (!("auto_youtube_upload" in $$source)) {
            /**
             * @member
             * @type {boolean}
             */
            this["auto_youtube_upload"] = false;
        }

        Object.assign(this, $$source);
    }

    /**
     * Creates a new Info instance from a string or object.
     * @param {any} [$$source = {}]
     * @returns {Info}
     */
    static createFrom($$source = {}) {
        let $$parsedSource = typeof $$source === 'string' ? JSON.parse($$source) : $$source;
        return new Info(/** @type {Partial<Info>} */($$parsedSource));
    }
}