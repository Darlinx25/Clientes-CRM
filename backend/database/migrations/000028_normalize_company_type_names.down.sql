-- Revertir normalización de nombres de tipos de empresa.
UPDATE company_types SET name = 'monotributo' WHERE name = 'Monotributo';
UPDATE company_types SET name = 'rural' WHERE name = 'Rural';
UPDATE company_types SET name = 'Administracion' WHERE name = 'Administración';
UPDATE company_types SET name = 'obra' WHERE name = 'Obra';