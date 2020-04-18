// (c) 2020, Ava Labs, Inc. All rights reserved.
// See the file LICENSE for licensing terms.

package avm_index

import (
	"encoding/json"

	"github.com/ava-labs/gecko/ids"
)

func (r *DBIndex) GetTxCount() (count int64, err error) {
	err = r.newDBSession("get_tx_count").
		Select("COUNT(1)").
		From("avm_transactions").
		Where("chain_id = ?", r.chainID.Bytes()).
		LoadOne(&count)
	return count, err
}

func (r *DBIndex) GetTx(id ids.ID) (*displayTx, error) {
	tx := &displayTx{}
	err := r.newDBSession("get_tx").
		Select("id", "json_serialization", "ingested_at").
		From("avm_transactions").
		Where("id = ?", id.Bytes()).
		Where("chain_id = ?", r.chainID.Bytes()).
		Limit(1).
		LoadOne(tx)
	return tx, err
}

func (r *DBIndex) GetTxs(params *ListTxParams) ([]*displayTx, error) {
	builder := params.Apply(r.newDBSession("get_txs").
		Select("id", "json_serialization", "ingested_at").
		From("avm_transactions").
		Where("chain_id = ?", r.chainID.Bytes()))

	txs := []*displayTx{}
	_, err := builder.Load(&txs)
	return txs, err
}

func (r *DBIndex) GetTxsForAddr(addr ids.ShortID, params *ListTxParams) ([]*displayTx, error) {
	builder := params.Apply(r.newDBSession("get_txs_for_address").
		SelectBySql(`
			SELECT id, json_serialization, ingested_at
			FROM avm_transactions
			LEFT JOIN avm_output_addresses AS oa1 ON avm_transactions.id = oa1.transaction_id
			LEFT JOIN avm_output_addresses AS oa2 ON avm_transactions.id = oa2.transaction_id
			WHERE
        avm_transactions.chain_id = ?
        AND
				oa1.output_index < oa2.output_index
				AND
				oa1.address = ?`, r.chainID.Bytes(), addr.Bytes()))

	txs := []*displayTx{}
	_, err := builder.Load(&txs)
	return txs, err
}

func (r *DBIndex) GetTxsForAsset(assetID ids.ID, params *ListTxParams) ([]json.RawMessage, error) {
	bytes := []json.RawMessage{}
	builder := params.Apply(r.newDBSession("get_txs_for_asset").
		SelectBySql(`
			SELECT avm_transactions.canonical_serialization
			FROM avm_transactions
			LEFT JOIN avm_output_addresses AS oa1 ON avm_transactions.id = oa1.transaction_id
			LEFT JOIN avm_output_addresses AS oa2 ON avm_transactions.id = oa2.transaction_id
			LEFT JOIN avm_outputs ON avm_outputs.transaction_id = oa1.transaction_id AND avm_outputs.output_index = oa1.output_index
			WHERE
        avm_outputs.asset_id = ?
        AND
        avm_transactions.chain_id = ?
        AND
				oa1.output_index < oa2.output_index`,
			assetID.Bytes, r.chainID.Bytes()))

	_, err := builder.Load(&bytes)
	return bytes, err

}

func (r *DBIndex) GetTXOsForAddr(addr ids.ShortID, params *ListTXOParams) ([]output, error) {
	builder := params.Apply(r.newDBSession("get_transaction").
		Select("*").
		From("avm_outputs").
		LeftJoin("avm_output_addresses", "avm_outputs.transaction_id = avm_output_addresses.transaction_id").
		LeftJoin("avm_transactions", "avm_transactions.id = avm_output_addresses.transaction_id").
		Where("avm_output_addresses.address = ?", addr.Bytes()).
		Where("avm_transactions.chain_id = ?", r.chainID.Bytes()))

	// TODO: Get addresses and add to outputs
	outputs := []output{}
	_, err := builder.Load(&outputs)
	return outputs, err
}

func (r *DBIndex) GetAssetCount() (count int64, err error) {
	err = r.newDBSession("get_asset_count").
		Select("COUNT(1)").
		From("avm_assets").
		Where("chain_id = ?", r.chainID.Bytes()).
		LoadOne(&count)
	return count, err
}

func (r *DBIndex) GetAssets(params *ListParams) ([]asset, error) {
	assets := []asset{}
	builder := params.Apply(r.newDBSession("get_assets").
		Select("*").
		From("avm_assets").
		Where("chain_id = ?", r.chainID.Bytes()))
	_, err := builder.Load(&assets)
	return assets, err
}

func (r *DBIndex) GetAsset(aliasOrID string) (asset, error) {
	a := asset{}
	query := r.newDBSession("get_asset").
		Select("*").
		From("avm_assets").
		Where("chain_id = ?", r.chainID.Bytes()).
		Limit(1)

	id, err := ids.FromString(aliasOrID)
	if err != nil {
		query = query.Where("alias = ?", aliasOrID)
	} else {
		query = query.Where("id = ?", id.Bytes())
	}

	err = query.LoadOne(&a)
	return a, err
}
