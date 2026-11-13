import {test} from 'node:test';
import assert from 'node:assert/strict';
import {splitLabels, templatePayload} from './form.js';

test('labels trim blanks and payload sends arrays without editing metadata', () => {
    assert.deepEqual(splitLabels('mircli, , crm,mircli'), ['mircli', 'crm']);
    const form = {code: 'receipt', name: 'Receipt', description: '', subject: 'Hi', body: 'Hi',
        params: {name: {default: '', required: false}}, systemsRaw: 'mircli, crm', channelsRaw: 'sms', order: 0, metadata: {keep: true}};
    assert.deepEqual(templatePayload(form), {code: 'receipt', name: 'Receipt', description: '', subject: 'Hi', body: 'Hi',
        params: form.params, order: 0, systems: ['mircli', 'crm'], channels: ['sms']});
    assert.equal(templatePayload({...form, order: undefined}).order, 1000);
    assert.deepEqual(splitLabels(''), []);
});
