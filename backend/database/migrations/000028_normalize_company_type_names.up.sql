-- Normalización de nombres de tipos de empresa:
-- mayúscula al inicio y consistente entre sí.
UPDATE company_types SET name = 'Monotributo' WHERE name = 'monotributo';
UPDATE company_types SET name = 'Rural' WHERE name = 'rural';
UPDATE company_types SET name = 'Administración' WHERE name = 'Administracion';
UPDATE company_types SET name = 'Obra' WHERE name = 'obra';