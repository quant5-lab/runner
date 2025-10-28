import { describe, it, expect } from 'vitest';
import { PineScriptTranspiler } from '../../src/pine/PineScriptTranspiler.js';

describe('PineScriptTranspiler - Parameter Shadowing Fix', () => {
  const transpiler = new PineScriptTranspiler();

  it('renames function parameter that shadows global input variable', async () => {
    const code = `
//@version=5
indicator("Test")
LWdilength = input(18, title="DMI Length")
adx(LWdilength, LWadxlength) =>
    value = LWdilength * 2
    value
result = adx(LWdilength, 20)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_LWdilength');
    expect(transpiled).not.toMatch(/const adx = \(LWdilength,/);
    expect(transpiled).toMatch(/const adx = \(_param_LWdilength/);
    expect(transpiled).toMatch(/const adx = \(_param_LWdilength, LWadxlength\) =>/);
    expect(transpiled).toMatch(/let value = _param_LWdilength \* 2/);
    expect(transpiled).toMatch(/adx\(.*LWdilength.*\)/);
  });

  it('keeps non-shadowing parameters unchanged', async () => {
    const code = `
indicator("Test")
length = input.int(14, title="Length")
calculate(period) =>
    period * 2
result = calculate(length)
plot(result)
    `;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).not.toContain('_param_period');
    expect(transpiled).toMatch(/const calculate = period =>/);
    expect(transpiled).toMatch(/return period \* 2/);
    expect(transpiled).toMatch(/calculate\(.*length.*\)/);
  }); it('handles multiple shadowing parameters in same function', async () => {
    const code = `
//@version=5
indicator("Test")
param1 = input(10)
param2 = input(20)
test(param1, param2) =>
    param1 + param2
result = test(param1, param2)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_param1');
    expect(transpiled).toContain('_param_param2');
    expect(transpiled).toMatch(/const test = \(_param_param1, _param_param2\) =>/);
    expect(transpiled).toMatch(/_param_param1 \+ _param_param2/);
    expect(transpiled).toMatch(/test\(.*param1.*param2.*\)/);
  });

  it('renames shadowing parameter throughout function body', async () => {
    const code = `
//@version=5
indicator("Test")
value = input(100)
process(value) =>
    temp = value * 2
    temp + value
result = process(value)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_value');
    expect(transpiled).toMatch(/_param_value \* 2/);
    expect(transpiled).toMatch(/temp \+ _param_value/);
    expect(transpiled).toMatch(/let temp = _param_value \* 2/);
    expect(transpiled).toMatch(/return temp \+ _param_value/);
    expect(transpiled).toMatch(/process\(.*value.*\)/);
    expect(transpiled).not.toMatch(/process\(.*_param_value.*\)/);
  });

  it('handles nested function scopes correctly', async () => {
    const code = `
//@version=5
indicator("Test")
outer = input(10)
level1(outer) =>
    level2(inner) =>
        inner * 2
    level2(outer)
result = level1(outer)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_outer');
    expect(transpiled).not.toContain('_param_inner');
    expect(transpiled).toMatch(/level2\(_param_outer\)/);
    expect(transpiled).toMatch(/const level2 = inner =>/);
    expect(transpiled).toMatch(/return inner \* 2/);
  });

  it('handles mixed shadowing and non-shadowing parameters', async () => {
    const code = `
//@version=5
indicator("Test")
length = input(10)
calculate(length, multiplier, offset) =>
    length * multiplier + offset
result = calculate(length, 2, 5)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_length');
    expect(transpiled).not.toContain('_param_multiplier');
    expect(transpiled).not.toContain('_param_offset');
    expect(transpiled).toMatch(/const calculate = \(_param_length, multiplier, offset\) =>/);
    expect(transpiled).toMatch(/_param_length \* multiplier \+ offset/);
  });

  it('handles shadowing parameter in complex expressions with ta functions', { timeout: 10000 }, async () => {
    const code = `
//@version=5
indicator("Test")
length = input(14)
dirmov(length) =>
    up = ta.change(high)
    down = -ta.change(low)
    ta.rma(up, length) + ta.rma(down, length)
result = dirmov(length)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_length');
    expect(transpiled).toMatch(/ta\.rma\(up, _param_length\)/);
    expect(transpiled).toMatch(/ta\.rma\(down, _param_length\)/);
    expect(transpiled).toMatch(/let up =/);
    expect(transpiled).toMatch(/let down =/);
    expect(transpiled).not.toMatch(/_param_up/);
    expect(transpiled).not.toMatch(/_param_down/);
  });

  it('handles triple-nested shadowing cascade', async () => {
    const code = `
//@version=5
indicator("Test")
value = input(100)
level1(value) =>
    level2(value) =>
        level3(value) =>
            value * 3
        level3(value * 2)
    level2(value)
result = level1(value)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_value');
    expect(transpiled).toMatch(/const level1 = _param_value =>/);
    expect(transpiled).toMatch(/const level2 = _param_value =>/);
    expect(transpiled).toMatch(/const level3 = _param_value =>/);
    expect(transpiled).toMatch(/return _param_value \* 3/);
  });

  it('handles shadowing parameter used in array indexing and conditionals', async () => {
    const code = `
//@version=5
indicator("Test")
index = input(0)
getValue(index) =>
    values = array.new_float(10, 0)
    array.get(values, index > 5 ? 5 : index)
result = getValue(index)
plot(result)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_index');
    expect(transpiled).toMatch(/_param_index > 5/);
    expect(transpiled).toMatch(/\? 5 : _param_index/);
    expect(transpiled).toMatch(/array\.get\(values, _param_index > 5 \? 5 : _param_index\)/);
  });

  it('handles function with multiple shadowing parameters and ta.rma calls', async () => {
    const code = `
//@version=5
indicator("Test")
LWdilength = input(18, title="DMI Length")
LWadxlength = input(20, title="ADX Length")
adx(LWdilength, LWadxlength) =>
    up = ta.change(high)
    down = -ta.change(low)
    plusDM = ta.rma(up, LWdilength)
    minusDM = ta.rma(down, LWdilength)
    adxValue = ta.rma(plusDM, LWadxlength)
    [adxValue, plusDM, minusDM]
[ADX, up, down] = adx(LWdilength, LWadxlength)
plot(ADX)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_LWdilength');
    expect(transpiled).toContain('_param_LWadxlength');
    expect(transpiled).toMatch(/const adx = \(_param_LWdilength, _param_LWadxlength\) =>/);
    expect(transpiled).toMatch(/ta\.rma\(up, _param_LWdilength\)/);
    expect(transpiled).toMatch(/ta\.rma\(down, _param_LWdilength\)/);
    expect(transpiled).toMatch(/ta\.rma\(plusDM, _param_LWadxlength\)/);
    expect(transpiled).toMatch(/adx\(.*LWdilength.*LWadxlength.*\)/);
    expect(transpiled).not.toMatch(/adx\(.*_param_LWdilength.*\)/);
    expect(transpiled).not.toMatch(/_param_up/);
    expect(transpiled).not.toMatch(/_param_down/);
    expect(transpiled).not.toMatch(/_param_plusDM/);
    expect(transpiled).not.toMatch(/_param_minusDM/);
  });

  it('handles shadowing parameter in tuple destructuring assignment', async () => {
    const code = `
//@version=5
indicator("Test")
len1 = input(10)
len2 = input(20)
calculate(len1, len2) =>
    sum = len1 + len2
    diff = len1 - len2
    [sum, diff]
[s, d] = calculate(len1, len2)
plot(s)
`;

    const transpiled = await transpiler.transpile(code);

    expect(transpiled).toContain('_param_len1');
    expect(transpiled).toContain('_param_len2');
    expect(transpiled).toMatch(/_param_len1 \+ _param_len2/);
    expect(transpiled).toMatch(/_param_len1 - _param_len2/);
    expect(transpiled).toMatch(/let sum = _param_len1 \+ _param_len2/);
    expect(transpiled).toMatch(/let diff = _param_len1 - _param_len2/);
  });
});
