ALTER TABLE zone_polygons ADD COLUMN floor TEXT CHECK (floor IN ('haut', 'bas'));
