// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import {Create as $Create} from "@wailsio/runtime";

/**
 * Represents the response body for the GetRecordStatus request.
 */
export class GetRecordStatusResponse {
    /**
     * Creates a new GetRecordStatusResponse instance.
     * @param {Partial<GetRecordStatusResponse>} [$$source = {}] - The source object to create the GetRecordStatusResponse.
     */
    constructor($$source = {}) {
        if (/** @type {any} */(false)) {
            /**
             * Whether the output is active
             * @member
             * @type {boolean | undefined}
             */
            this["outputActive"] = false;
        }
        if (/** @type {any} */(false)) {
            /**
             * Number of bytes sent by the output
             * @member
             * @type {number | undefined}
             */
            this["outputBytes"] = 0;
        }
        if (/** @type {any} */(false)) {
            /**
             * Current duration in milliseconds for the output
             * @member
             * @type {number | undefined}
             */
            this["outputDuration"] = 0;
        }
        if (/** @type {any} */(false)) {
            /**
             * Whether the output is paused
             * @member
             * @type {boolean | undefined}
             */
            this["outputPaused"] = false;
        }
        if (/** @type {any} */(false)) {
            /**
             * Current formatted timecode string for the output
             * @member
             * @type {string | undefined}
             */
            this["outputTimecode"] = "";
        }

        Object.assign(this, $$source);
    }

    /**
     * Creates a new GetRecordStatusResponse instance from a string or object.
     * @param {any} [$$source = {}]
     * @returns {GetRecordStatusResponse}
     */
    static createFrom($$source = {}) {
        let $$parsedSource = typeof $$source === 'string' ? JSON.parse($$source) : $$source;
        return new GetRecordStatusResponse(/** @type {Partial<GetRecordStatusResponse>} */($$parsedSource));
    }
}
