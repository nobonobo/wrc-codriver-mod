// @ts-check
// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import {Call as $Call, Create as $Create} from "@wailsio/runtime";

// eslint-disable-next-line @typescript-eslint/ban-ts-comment
// @ts-ignore: Unused imports
import * as $models from "./models.js";

/**
 * @returns {Promise<$models.Settings> & { cancel(): void }}
 */
export function Get() {
    let $resultPromise = /** @type {any} */($Call.ByID(1895040484));
    let $typingPromise = /** @type {any} */($resultPromise.then(($result) => {
        return $$createType0($result);
    }));
    $typingPromise.cancel = $resultPromise.cancel.bind($resultPromise);
    return $typingPromise;
}

/**
 * @param {$models.Settings} settings
 * @returns {Promise<void> & { cancel(): void }}
 */
export function Update(settings) {
    let $resultPromise = /** @type {any} */($Call.ByID(799263641, settings));
    return $resultPromise;
}

// Private type creation functions
const $$createType0 = $models.Settings.createFrom;
