const fs = require('node:fs');
const path = require('node:path');
const crypto = require('node:crypto');
const root = path.resolve(__dirname, '../../data/iran-geography');
const read = name => JSON.parse(fs.readFileSync(path.join(root, 'upstream', name), 'utf8'));
const provinces = read('provinces.json');
const cities = read('cities-filtered.json');
const counties = new Map(read('counties.json').map(c => [c.id, c]));
const manifest = read('manifest.json');
const key = value => value.normalize('NFC').replace(/[يى]/g, 'ی').replace(/ك/g, 'ک').replace(/[\s\u200c\u200d\u0640\u064b-\u065f]/g, '');
const displayName = value => value.normalize('NFC').replace(/[يى]/g, 'ی').replace(/ك/g, 'ک').trim().replace(/\s+/g, ' ');
const count = name => manifest.generatedDatasets.find(d => d.name === name)?.rowCount;
if (provinces.length !== 31 || provinces.length !== count('provinces') || cities.length !== count('cities-filtered')) throw Error('Upstream count mismatch');
const provinceIDs = new Set(provinces.map(p => p.id));
const used = new Set();
const duplicates = new Map();
for (const c of cities) {
  if (!Number.isSafeInteger(c.id) || !provinceIDs.has(c.province_id) || used.has(c.id) || !c.name?.trim()) throw Error('Invalid upstream city');
  used.add(c.id);
  const id = c.province_id + ':' + key(c.name);
  duplicates.set(id, (duplicates.get(id) ?? 0) + 1);
}
const disambiguations = [];
const directory = provinces.map(p => ({
  id: 'ir-p-' + p.id, name: displayName(p.name), active: true,
  cities: cities.filter(c => c.province_id === p.id).map(c => {
    let name = displayName(c.name);
    if (duplicates.get(c.province_id + ':' + key(c.name)) > 1) {
      const county = counties.get(c.county_id);
      if (!county || county.province_id !== c.province_id) throw Error('Missing county for duplicate city');
      name += ' (' + displayName(county.name) + ')';
      disambiguations.push({city_id:c.id, original_name:c.name, display_name:name, county_id:c.county_id});
    }
    return {id:'ir-c-' + c.id, name, active:true};
  }),
}));
for (const p of directory) {
  if (!p.cities.length || new Set(p.cities.map(c => key(c.name))).size !== p.cities.length) throw Error('Ambiguous normalized city names');
}
const source = {
  repository:'https://github.com/sajaddp/list-of-cities-in-Iran',
  commit:'ace21bb251f96ced350bbc01541acffd71121feb',
  version:manifest.datasetVersion, official_year:manifest.officialSourceYear,
  city_dataset:'cities-filtered', province_count:provinces.length, city_count:cities.length,
};
const seed = {dataset_id:`iran-geography-${source.official_year}-v${source.version}`, source, provinces:directory};
const encoded = JSON.stringify(seed, null, 2) + '\n';
fs.writeFileSync(path.join(root, 'iran-provinces-cities.json'), encoded);
const hashes = Object.fromEntries(['provinces.json','cities-filtered.json','counties.json','manifest.json','license.txt'].map(name => [name, crypto.createHash('sha256').update(fs.readFileSync(path.join(root,'upstream',name))).digest('hex')]));
fs.writeFileSync(path.join(root, 'provenance.json'), JSON.stringify({
  ...source, retrieved_on:'2026-11-27', license:'GPL-3.0',
  official_workbook_sha256:manifest.source.sha256,
  upstream_sha256:hashes,
  dataset_sha256:crypto.createHash('sha256').update(encoded).digest('hex'),
  excluded_urban_zones:count('urban-zones'), name_disambiguations:disambiguations,
}, null, 2) + '\n');
console.log(JSON.stringify({provinces:provinces.length,cities:cities.length,disambiguated:disambiguations.length,compact_bytes:Buffer.byteLength(JSON.stringify(seed.provinces))}));
