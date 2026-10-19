import assert from "node:assert/strict";
import test from "node:test";

import { mergeWorkflowFieldMappings, normalizeWorkflowFieldMappings } from "../src/stores/config-workflow-fields.ts";
import { workflowFieldChoiceValues, workflowFieldValueError } from "../src/lib/model-capabilities-workflow.ts";

const appFields = [
    { nodeId: "51", fieldName: "image", source: "referenceImage", sourceAutomatic: true, imageOrder: 1, sourceIndex: 0, required: false, fieldValue: "example.png", emptyValue: "None", label: "image1（图1）" },
    { nodeId: "49", fieldName: "image", source: "referenceImage", sourceAutomatic: true, imageOrder: 2, sourceIndex: 1, required: false, fieldValue: "None" },
    { nodeId: "48", fieldName: "audio", source: "referenceAudio", sourceAutomatic: true, sourceIndex: 0, required: false, fieldValue: "None" },
    { nodeId: "27", fieldName: "video", source: "referenceVideo", sourceAutomatic: true, sourceIndex: 0, required: false, fieldValue: "" },
];

test("App fields retain public order, optional media and empty values after persistence", () => {
    const fields = normalizeWorkflowFieldMappings(JSON.parse(JSON.stringify(normalizeWorkflowFieldMappings(appFields, "video"))), "video");
    assert.deepEqual(fields.map((field) => field.nodeId), ["51", "49", "48", "27"]);
    assert.deepEqual(fields.map((field) => field.required), [false, false, false, false]);
    assert.deepEqual(fields.map((field) => field.fieldValue), ["example.png", "None", "None", ""]);
    assert.equal(fields[0].label, "image1（图1）");
    assert.equal(fields[0].emptyValue, "None");
});

test("refetch rebuilds automatic App media slots in incoming order", () => {
    const previous = [
        { ...appFields[1], imageOrder: 1, sourceIndex: 0, required: true },
        { ...appFields[0], imageOrder: 2, sourceIndex: 1, required: true },
    ];
    const fields = mergeWorkflowFieldMappings(previous, appFields, "video", "app");
    assert.deepEqual(fields.map((field) => field.nodeId), ["51", "49", "48", "27"]);
    assert.equal(fields[0].imageOrder, 1);
    assert.equal(fields[1].imageOrder, 2);
    assert.equal(fields[0].required, false);
    assert.equal(fields[1].required, false);
});

test("refetch preserves manual media policies and Workflow behavior", () => {
    const manual = { ...appFields[0], sourceAutomatic: false, imageOrder: 3, sourceIndex: 2, required: true, enabled: false, label: "Custom" };
    const fields = mergeWorkflowFieldMappings([manual], appFields, "video", "app");
    assert.equal(fields[0].imageOrder, 3);
    assert.equal(fields[0].required, true);
    assert.equal(fields[0].enabled, false);
    assert.equal(fields[0].label, "Custom");
    const workflow = mergeWorkflowFieldMappings([{ ...manual, sourceAutomatic: true }], appFields, "video", "workflow");
    assert.equal(workflow[0].imageOrder, 3);
    assert.equal(workflow[0].required, true);
});

test("App options and numeric bounds survive normalization and validate actual values", () => {
    const fields = normalizeWorkflowFieldMappings([
        { nodeId: "252", fieldName: "aspect_ratio", source: "", fieldType: "LIST", fieldValue: "16:9 (Widescreen)", options: ["1:1 (Square)", "16:9 (Widescreen)"] },
        { nodeId: "252", fieldName: "megapixels", source: "", fieldType: "FLOAT", fieldValue: "0.5", min: 0.1, max: 16, step: 0.1 },
        { nodeId: "1", fieldName: "model", source: "", fieldType: "LIST", fieldValue: "fast", options: [{ value: "fast", label: "快速" }] },
    ], "video");
    assert.deepEqual(workflowFieldChoiceValues(fields[0]), ["1:1 (Square)", "16:9 (Widescreen)"]);
    assert.equal(workflowFieldValueError(fields[0], "16:9 (Widescreen)"), "");
    assert.notEqual(workflowFieldValueError(fields[0], "9:16"), "");
    assert.equal(workflowFieldValueError(fields[1], "0.5"), "");
    assert.notEqual(workflowFieldValueError(fields[1], "0.55"), "");
    assert.deepEqual(workflowFieldChoiceValues(fields[2]), ["fast"]);
});
